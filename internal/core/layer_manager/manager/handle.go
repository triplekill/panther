package manager

/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"go.uber.org/zap"

	analysisoperations "github.com/panther-labs/panther/api/gateway/analysis/client/operations"
	"github.com/panther-labs/panther/api/gateway/analysis/models"
)

const (
	layerPath        = "python/lib/python3.7/site-packages/"
	layerRuntime     = "python3.7"
	globalModuleName = "panther"
)

var (
	globalLayerName  = aws.String(os.Getenv("GLOBAL_LAYER"))
	policyEngineName = aws.String(os.Getenv("POLICY_ENGINE"))
	ruleEngineName   = aws.String(os.Getenv("RULE_ENGINE"))
)

// UpdateLayer rebuilds and publishes the layer for the given analysis type.
// Currently global is the only supported analysis type.
func UpdateLayer(analysisType string) error {
	if analysisType != string(models.AnalysisTypeGLOBAL) {
		zap.L().Warn("unsupported analysis type", zap.String("type", analysisType))
		// When we add support for policies/rules, we can use this variable to control which layers are re-created
		// and from which sources. We can either have entirely separate paths for these, or have some sort of config
		// stored that records the different names, paths, etc. mapped to the different analysis types.
		return errors.New("cannot build layer for unsupported analysisType " + analysisType)
	}

	newLayer, err := buildLayer()
	if err != nil {
		return err
	}

	layerArn, layerVersionArn, err := publishLayer(newLayer)
	if err != nil {
		return err
	}

	// For policy/rule layers, only do one of these
	err = updateLambda(policyEngineName, layerArn, layerVersionArn)
	if err != nil {
		return err
	}

	return updateLambda(ruleEngineName, layerArn, layerVersionArn)
}

// buildLayer looks up the required analyses and from them constructs the zip archive that defines the layer
func buildLayer() ([]byte, error) {
	zap.L().Debug("building lambda layer")
	// TODO: talk to the analysis-api GetEnabledPolicies endpoint and build the layer for policies/rules
	// be sure to have a means of differentiating the resource/log type of each policy/rule

	// When multiple globals are supported, this can be updated to get a list
	global, err := analysisClient.Operations.GetGlobal(&analysisoperations.GetGlobalParams{
		GlobalID:   globalModuleName,
		HTTPClient: httpClient,
	})
	if err != nil {
		return nil, err
	}
	return packageLayer(map[string]string{globalModuleName: string(global.Payload.Body)})
}

// packageLayer takes a mapping of filenames to function bodies and constructs a zip archive with the file structure
// that AWS is expecting.
func packageLayer(analyses map[string]string) ([]byte, error) {
	zap.L().Debug("packaging lambda layer")
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for id, body := range analyses {
		f, err := w.Create(layerPath + id + ".py")
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(body))
		if err != nil {
			return nil, err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// publishLayer takes a zip file and publishes it as a new lambda layer. It returns both the layer ARN and the layer
// ARN with version, for simplicity's sake.
func publishLayer(layerBody []byte) (*string, *string, error) {
	zap.L().Debug("publishing lambda layer")
	layer, err := lambdaClient.PublishLayerVersion(&lambda.PublishLayerVersionInput{
		CompatibleRuntimes: []*string{aws.String(layerRuntime)},
		Content: &lambda.LayerVersionContentInput{
			ZipFile: layerBody,
		},
		Description: aws.String("The panther engine global helper layer."),
		LayerName:   globalLayerName,
	})
	if err != nil {
		return nil, nil, err
	}
	return layer.LayerArn, layer.LayerVersionArn, nil
}

// updateLambda updates the function configuration of a given lambda to include the specified lambda layer.
// The layer is updated to the given version if it is already present.
func updateLambda(lambdaName, layerArn, layerVersionArn *string) error {
	zap.L().Debug("updating lambda function with new layer", zap.String("lambda", *lambdaName), zap.String("layer", *layerVersionArn))
	// Lambda does not let you update just one layer on a lambda, you must specify the name of each desired layer so
	// we start by listing what layers are already present to preserve them.
	oldLayers, err := lambdaClient.GetFunctionConfiguration(&lambda.GetFunctionConfigurationInput{
		FunctionName: lambdaName,
	})
	if err != nil {
		return err
	}

	// Replace the layer we want to update with the new layer
	newLayers := make([]*string, len(oldLayers.Layers))
	replaced := false
	for i, layer := range oldLayers.Layers {
		if strings.HasPrefix(*layer.Arn, *layerArn) {
			newLayers[i] = layerVersionArn
			replaced = true
		} else {
			newLayers[i] = layer.Arn
		}
	}

	// Handle the case where we are not updating an existing layer
	if !replaced {
		zap.L().Debug("no lambda layer to replace")
		newLayers = append(newLayers, layerVersionArn)
	}

	// Update the lambda function. This operation may take 1-3 seconds.
	_, err = lambdaClient.UpdateFunctionConfiguration(&lambda.UpdateFunctionConfigurationInput{
		FunctionName: lambdaName,
		Layers:       newLayers,
	})

	return err
}