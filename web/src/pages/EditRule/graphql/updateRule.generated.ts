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

/* eslint-disable import/order, import/no-duplicates, @typescript-eslint/no-unused-vars */

import * as Types from '../../../../__generated__/schema';

import gql from 'graphql-tag';
import * as ApolloReactCommon from '@apollo/client';
import * as ApolloReactHooks from '@apollo/client';

export type UpdateRuleVariables = {
  input: Types.CreateOrModifyRuleInput;
};

export type UpdateRule = {
  updateRule?: Types.Maybe<
    Pick<
      Types.RuleDetails,
      | 'description'
      | 'displayName'
      | 'enabled'
      | 'id'
      | 'reference'
      | 'logTypes'
      | 'runbook'
      | 'severity'
      | 'dedupPeriodMinutes'
      | 'tags'
      | 'body'
    > & {
      tests?: Types.Maybe<
        Array<
          Types.Maybe<
            Pick<Types.PolicyUnitTest, 'expectedResult' | 'name' | 'resource' | 'resourceType'>
          >
        >
      >;
    }
  >;
};

export const UpdateRuleDocument = gql`
  mutation UpdateRule($input: CreateOrModifyRuleInput!) {
    updateRule(input: $input) {
      description
      displayName
      enabled
      id
      reference
      logTypes
      runbook
      severity
      dedupPeriodMinutes
      tags
      body
      tests {
        expectedResult
        name
        resource
        resourceType
      }
    }
  }
`;
export type UpdateRuleMutationFn = ApolloReactCommon.MutationFunction<
  UpdateRule,
  UpdateRuleVariables
>;

/**
 * __useUpdateRule__
 *
 * To run a mutation, you first call `useUpdateRule` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateRule` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateRule, { data, loading, error }] = useUpdateRule({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateRule(
  baseOptions?: ApolloReactHooks.MutationHookOptions<UpdateRule, UpdateRuleVariables>
) {
  return ApolloReactHooks.useMutation<UpdateRule, UpdateRuleVariables>(
    UpdateRuleDocument,
    baseOptions
  );
}
export type UpdateRuleHookResult = ReturnType<typeof useUpdateRule>;
export type UpdateRuleMutationResult = ApolloReactCommon.MutationResult<UpdateRule>;
export type UpdateRuleMutationOptions = ApolloReactCommon.BaseMutationOptions<
  UpdateRule,
  UpdateRuleVariables
>;
