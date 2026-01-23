import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  JSON: { input: any; output: any; }
  Time: { input: string; output: string; }
};

export type Me = {
  __typename?: 'Me';
  email: Maybe<Scalars['String']['output']>;
  name: Maybe<Scalars['String']['output']>;
  orgID: Scalars['ID']['output'];
  role: Scalars['String']['output'];
  userID: Scalars['ID']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  approveRequest: Request;
  createRequest: Request;
  /** Create a workflow template (admin only). */
  createWorkflowTemplate: WorkflowTemplate;
  rejectRequest: Request;
  resubmitRequest: Request;
  returnRequest: Request;
  submitRequest: Request;
};


export type MutationApproveRequestArgs = {
  id: Scalars['ID']['input'];
};


export type MutationCreateRequestArgs = {
  title: Scalars['String']['input'];
  workflowTemplateID: InputMaybe<Scalars['ID']['input']>;
};


export type MutationCreateWorkflowTemplateArgs = {
  definition: Scalars['JSON']['input'];
  description: Scalars['String']['input'];
  name: Scalars['String']['input'];
};


export type MutationRejectRequestArgs = {
  id: Scalars['ID']['input'];
  reason: Scalars['String']['input'];
};


export type MutationResubmitRequestArgs = {
  id: Scalars['ID']['input'];
};


export type MutationReturnRequestArgs = {
  id: Scalars['ID']['input'];
  reason: Scalars['String']['input'];
};


export type MutationSubmitRequestArgs = {
  id: Scalars['ID']['input'];
};

export type Query = {
  __typename?: 'Query';
  /** Returns the current logged-in user (viewer). */
  me: Me;
  /** Health-check style ping. Requires login because the endpoint is guarded. */
  ping: Scalars['String']['output'];
  /** Get a single request within the viewer's org. */
  request: Maybe<Request>;
  /** List requests within the viewer's org. */
  requests: Array<Request>;
  /** List workflow templates within the viewer's org. */
  workflowTemplates: Array<WorkflowTemplate>;
};


export type QueryRequestArgs = {
  id: Scalars['ID']['input'];
};


export type QueryRequestsArgs = {
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryWorkflowTemplatesArgs = {
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
};

export type Request = {
  __typename?: 'Request';
  auditTrail: Array<RequestAudit>;
  createdAt: Scalars['Time']['output'];
  createdByUserID: Maybe<Scalars['ID']['output']>;
  decidedAt: Maybe<Scalars['Time']['output']>;
  decidedByUserID: Maybe<Scalars['ID']['output']>;
  id: Scalars['ID']['output'];
  orgID: Scalars['ID']['output'];
  status: Scalars['String']['output'];
  steps: Array<RequestStep>;
  submittedAt: Maybe<Scalars['Time']['output']>;
  title: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type RequestAudit = {
  __typename?: 'RequestAudit';
  action: Scalars['String']['output'];
  actorUserID: Maybe<Scalars['ID']['output']>;
  data: Scalars['JSON']['output'];
  id: Scalars['ID']['output'];
  occurredAt: Scalars['Time']['output'];
};

export type RequestStep = {
  __typename?: 'RequestStep';
  assignedToUserID: Maybe<Scalars['ID']['output']>;
  label: Scalars['String']['output'];
  status: Scalars['String']['output'];
  stepIndex: Scalars['Int']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type WorkflowTemplate = {
  __typename?: 'WorkflowTemplate';
  createdAt: Scalars['Time']['output'];
  createdByUserID: Maybe<Scalars['ID']['output']>;
  definition: Scalars['JSON']['output'];
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  orgID: Scalars['ID']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'Query', me: { __typename?: 'Me', userID: string, orgID: string, role: string, name: string | null, email: string | null } };

export type RequestQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type RequestQuery = { __typename?: 'Query', request: { __typename?: 'Request', id: string, orgID: string, title: string, status: string, createdByUserID: string | null, decidedByUserID: string | null, createdAt: string, updatedAt: string, submittedAt: string | null, decidedAt: string | null, steps: Array<{ __typename?: 'RequestStep', stepIndex: number, label: string, status: string, assignedToUserID: string | null, updatedAt: string }>, auditTrail: Array<{ __typename?: 'RequestAudit', id: string, actorUserID: string | null, action: string, data: any, occurredAt: string }> } | null };

export type RequestsQueryVariables = Exact<{
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
}>;


export type RequestsQuery = { __typename?: 'Query', requests: Array<{ __typename?: 'Request', id: string, title: string, status: string, updatedAt: string }> };


export const MeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Me"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"me"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"userID"}},{"kind":"Field","name":{"kind":"Name","value":"orgID"}},{"kind":"Field","name":{"kind":"Name","value":"role"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}}]}}]} as unknown as DocumentNode<MeQuery, MeQueryVariables>;
export const RequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Request"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"request"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"orgID"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdByUserID"}},{"kind":"Field","name":{"kind":"Name","value":"decidedByUserID"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}},{"kind":"Field","name":{"kind":"Name","value":"submittedAt"}},{"kind":"Field","name":{"kind":"Name","value":"decidedAt"}},{"kind":"Field","name":{"kind":"Name","value":"steps"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"stepIndex"}},{"kind":"Field","name":{"kind":"Name","value":"label"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"assignedToUserID"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"auditTrail"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"actorUserID"}},{"kind":"Field","name":{"kind":"Name","value":"action"}},{"kind":"Field","name":{"kind":"Name","value":"data"}},{"kind":"Field","name":{"kind":"Name","value":"occurredAt"}}]}}]}}]}}]} as unknown as DocumentNode<RequestQuery, RequestQueryVariables>;
export const RequestsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Requests"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"50"}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"0"}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"requests"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<RequestsQuery, RequestsQueryVariables>;