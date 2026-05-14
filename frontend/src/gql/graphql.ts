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
  Upload: { input: any; output: any; }
};

export type Attachment = {
  __typename?: 'Attachment';
  contentType: Scalars['String']['output'];
  createdAt: Scalars['Time']['output'];
  filename: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  requestID: Scalars['ID']['output'];
  size: Scalars['Int']['output'];
  uploadedByUserID: Maybe<Scalars['ID']['output']>;
};

export type AvgTimeToApprovalResult = {
  __typename?: 'AvgTimeToApprovalResult';
  avgSeconds: Scalars['Float']['output'];
  sampleCount: Scalars['Int']['output'];
};

export type Comment = {
  __typename?: 'Comment';
  content: Scalars['String']['output'];
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  requestID: Scalars['ID']['output'];
  userID: Scalars['ID']['output'];
};

export type CountRequestsByMonthRow = {
  __typename?: 'CountRequestsByMonthRow';
  count: Scalars['Int']['output'];
  month: Scalars['Time']['output'];
};

export type CountRequestsByStatusRow = {
  __typename?: 'CountRequestsByStatusRow';
  count: Scalars['Int']['output'];
  status: Scalars['String']['output'];
};

export type DashboardSummaryResult = {
  __typename?: 'DashboardSummaryResult';
  approvedCount: Scalars['Int']['output'];
  avgApprovalSeconds: Scalars['Float']['output'];
  draftCount: Scalars['Int']['output'];
  rejectedCount: Scalars['Int']['output'];
  submittedCount: Scalars['Int']['output'];
  totalCount: Scalars['Int']['output'];
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
  addComment: Comment;
  approveRequest: Request;
  createRequest: Request;
  /** Create a workflow template (admin only). */
  createWorkflowTemplate: WorkflowTemplate;
  rejectRequest: Request;
  resubmitRequest: Request;
  returnRequest: Request;
  submitRequest: Request;
  /** Upload an attachment to a request */
  uploadAttachment: Attachment;
};


export type MutationAddCommentArgs = {
  content: Scalars['String']['input'];
  requestID: Scalars['ID']['input'];
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


export type MutationUploadAttachmentArgs = {
  file: Scalars['Upload']['input'];
  requestID: Scalars['ID']['input'];
};

export type Query = {
  __typename?: 'Query';
  /** Get a presigned download URL for an attachment. URL expires in 15 minutes. */
  attachmentDownloadURL: Scalars['String']['output'];
  /** Search audit logs with optional filters (admin/approver only). */
  auditLogs: Array<RequestAudit>;
  /** Get average time to approval (in seconds) for approved requests. */
  avgTimeToApproval: AvgTimeToApprovalResult;
  /** List comments */
  comments: Array<Comment>;
  /** Count requests grouped by month (optionally filtered by date range). */
  countRequestsByMonth: Array<CountRequestsByMonthRow>;
  /** Count requests grouped by status. */
  countRequestsByStatus: Array<CountRequestsByStatusRow>;
  /** Get comprehensive dashboard summary in a single query. */
  dashboardSummary: DashboardSummaryResult;
  /** Returns the current logged-in user (viewer). */
  me: Me;
  /** Health-check style ping. Requires login because the endpoint is guarded. */
  ping: Scalars['String']['output'];
  /** Get a single request within the viewer's org. */
  request: Maybe<Request>;
  /** List requests within the viewer's org. */
  requests: Array<Request>;
  /** Search requests within the viewer's org with optional filters. */
  searchRequests: Array<Request>;
  /** List workflow templates within the viewer's org. */
  workflowTemplates: Array<WorkflowTemplate>;
};


export type QueryAttachmentDownloadUrlArgs = {
  attachmentID: Scalars['ID']['input'];
  requestID: Scalars['ID']['input'];
};


export type QueryAuditLogsArgs = {
  action: InputMaybe<Scalars['String']['input']>;
  actorUserID: InputMaybe<Scalars['ID']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  occurredAtEnd: InputMaybe<Scalars['Time']['input']>;
  occurredAtStart: InputMaybe<Scalars['Time']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
  requestID: InputMaybe<Scalars['ID']['input']>;
};


export type QueryCommentsArgs = {
  requestID: Scalars['ID']['input'];
};


export type QueryCountRequestsByMonthArgs = {
  endDate: InputMaybe<Scalars['Time']['input']>;
  startDate: InputMaybe<Scalars['Time']['input']>;
};


export type QueryRequestArgs = {
  id: Scalars['ID']['input'];
};


export type QueryRequestsArgs = {
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
};


export type QuerySearchRequestsArgs = {
  createdAtEnd: InputMaybe<Scalars['Time']['input']>;
  createdAtStart: InputMaybe<Scalars['Time']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
  status: InputMaybe<Scalars['String']['input']>;
  title: InputMaybe<Scalars['String']['input']>;
};


export type QueryWorkflowTemplatesArgs = {
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
};

export type Request = {
  __typename?: 'Request';
  attachments: Array<Attachment>;
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
  submitter: Maybe<User>;
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

export type User = {
  __typename?: 'User';
  email: Maybe<Scalars['String']['output']>;
  name: Maybe<Scalars['String']['output']>;
  orgID: Scalars['ID']['output'];
  userID: Scalars['ID']['output'];
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

export type CreateRequestMutationVariables = Exact<{
  title: Scalars['String']['input'];
  workflowTemplateID: InputMaybe<Scalars['ID']['input']>;
}>;


export type CreateRequestMutation = { __typename?: 'Mutation', createRequest: { __typename?: 'Request', id: string, title: string, status: string, updatedAt: string } };

export type SubmitRequestMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type SubmitRequestMutation = { __typename?: 'Mutation', submitRequest: { __typename?: 'Request', id: string, status: string, updatedAt: string } };

export type ApproveRequestMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type ApproveRequestMutation = { __typename?: 'Mutation', approveRequest: { __typename?: 'Request', id: string, status: string, updatedAt: string } };

export type RejectRequestMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  reason: Scalars['String']['input'];
}>;


export type RejectRequestMutation = { __typename?: 'Mutation', rejectRequest: { __typename?: 'Request', id: string, status: string, updatedAt: string } };

export type ReturnRequestMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  reason: Scalars['String']['input'];
}>;


export type ReturnRequestMutation = { __typename?: 'Mutation', returnRequest: { __typename?: 'Request', id: string, status: string, updatedAt: string } };

export type ResubmitRequestMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type ResubmitRequestMutation = { __typename?: 'Mutation', resubmitRequest: { __typename?: 'Request', id: string, status: string, updatedAt: string } };

export type AttachmentDownloadUrlQueryVariables = Exact<{
  requestID: Scalars['ID']['input'];
  attachmentID: Scalars['ID']['input'];
}>;


export type AttachmentDownloadUrlQuery = { __typename?: 'Query', attachmentDownloadURL: string };

export type AuditLogsQueryVariables = Exact<{
  requestID: InputMaybe<Scalars['ID']['input']>;
  actorUserID: InputMaybe<Scalars['ID']['input']>;
  action: InputMaybe<Scalars['String']['input']>;
  occurredAtStart: InputMaybe<Scalars['Time']['input']>;
  occurredAtEnd: InputMaybe<Scalars['Time']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
}>;


export type AuditLogsQuery = { __typename?: 'Query', auditLogs: Array<{ __typename?: 'RequestAudit', id: string, actorUserID: string | null, action: string, data: any, occurredAt: string }> };

export type RequestCommentsQueryVariables = Exact<{
  requestID: Scalars['ID']['input'];
}>;


export type RequestCommentsQuery = { __typename?: 'Query', comments: Array<{ __typename?: 'Comment', id: string, requestID: string, userID: string, content: string, createdAt: string }> };

export type AddCommentMutationVariables = Exact<{
  requestID: Scalars['ID']['input'];
  content: Scalars['String']['input'];
}>;


export type AddCommentMutation = { __typename?: 'Mutation', addComment: { __typename?: 'Comment', id: string, requestID: string, userID: string, content: string, createdAt: string } };

export type DashboardSummaryQueryVariables = Exact<{ [key: string]: never; }>;


export type DashboardSummaryQuery = { __typename?: 'Query', dashboardSummary: { __typename?: 'DashboardSummaryResult', draftCount: number, submittedCount: number, approvedCount: number, rejectedCount: number, totalCount: number, avgApprovalSeconds: number } };

export type CountRequestsByStatusQueryVariables = Exact<{ [key: string]: never; }>;


export type CountRequestsByStatusQuery = { __typename?: 'Query', countRequestsByStatus: Array<{ __typename?: 'CountRequestsByStatusRow', status: string, count: number }> };

export type CountRequestsByMonthQueryVariables = Exact<{
  startDate: InputMaybe<Scalars['Time']['input']>;
  endDate: InputMaybe<Scalars['Time']['input']>;
}>;


export type CountRequestsByMonthQuery = { __typename?: 'Query', countRequestsByMonth: Array<{ __typename?: 'CountRequestsByMonthRow', month: string, count: number }> };

export type AvgTimeToApprovalQueryVariables = Exact<{ [key: string]: never; }>;


export type AvgTimeToApprovalQuery = { __typename?: 'Query', avgTimeToApproval: { __typename?: 'AvgTimeToApprovalResult', avgSeconds: number, sampleCount: number } };

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'Query', me: { __typename?: 'Me', userID: string, orgID: string, role: string, name: string | null, email: string | null } };

export type RequestQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type RequestQuery = { __typename?: 'Query', request: { __typename?: 'Request', id: string, orgID: string, title: string, status: string, createdByUserID: string | null, decidedByUserID: string | null, createdAt: string, updatedAt: string, submittedAt: string | null, decidedAt: string | null, steps: Array<{ __typename?: 'RequestStep', stepIndex: number, label: string, status: string, assignedToUserID: string | null, updatedAt: string }>, auditTrail: Array<{ __typename?: 'RequestAudit', id: string, actorUserID: string | null, action: string, data: any, occurredAt: string }>, attachments: Array<{ __typename?: 'Attachment', id: string, requestID: string, filename: string, contentType: string, size: number, uploadedByUserID: string | null, createdAt: string }> } | null };

export type RequestsQueryVariables = Exact<{
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
}>;


export type RequestsQuery = { __typename?: 'Query', requests: Array<{ __typename?: 'Request', id: string, title: string, status: string, updatedAt: string }> };

export type SearchRequestsQueryVariables = Exact<{
  title: InputMaybe<Scalars['String']['input']>;
  status: InputMaybe<Scalars['String']['input']>;
  createdAtStart: InputMaybe<Scalars['Time']['input']>;
  createdAtEnd: InputMaybe<Scalars['Time']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
}>;


export type SearchRequestsQuery = { __typename?: 'Query', searchRequests: Array<{ __typename?: 'Request', id: string, orgID: string, title: string, status: string, createdByUserID: string | null, decidedByUserID: string | null, createdAt: string, updatedAt: string, submittedAt: string | null, decidedAt: string | null }> };

export type UploadAttachmentMutationVariables = Exact<{
  requestID: Scalars['ID']['input'];
  file: Scalars['Upload']['input'];
}>;


export type UploadAttachmentMutation = { __typename?: 'Mutation', uploadAttachment: { __typename?: 'Attachment', id: string, requestID: string, filename: string, contentType: string, size: number, createdAt: string } };


export const CreateRequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateRequest"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"title"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"workflowTemplateID"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createRequest"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"title"},"value":{"kind":"Variable","name":{"kind":"Name","value":"title"}}},{"kind":"Argument","name":{"kind":"Name","value":"workflowTemplateID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"workflowTemplateID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<CreateRequestMutation, CreateRequestMutationVariables>;
export const SubmitRequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SubmitRequest"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"submitRequest"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<SubmitRequestMutation, SubmitRequestMutationVariables>;
export const ApproveRequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ApproveRequest"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"approveRequest"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<ApproveRequestMutation, ApproveRequestMutationVariables>;
export const RejectRequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RejectRequest"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"reason"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"rejectRequest"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}},{"kind":"Argument","name":{"kind":"Name","value":"reason"},"value":{"kind":"Variable","name":{"kind":"Name","value":"reason"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<RejectRequestMutation, RejectRequestMutationVariables>;
export const ReturnRequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ReturnRequest"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"reason"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"returnRequest"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}},{"kind":"Argument","name":{"kind":"Name","value":"reason"},"value":{"kind":"Variable","name":{"kind":"Name","value":"reason"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<ReturnRequestMutation, ReturnRequestMutationVariables>;
export const ResubmitRequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ResubmitRequest"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"resubmitRequest"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<ResubmitRequestMutation, ResubmitRequestMutationVariables>;
export const AttachmentDownloadUrlDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AttachmentDownloadURL"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"attachmentID"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"attachmentDownloadURL"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"requestID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}}},{"kind":"Argument","name":{"kind":"Name","value":"attachmentID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"attachmentID"}}}]}]}}]} as unknown as DocumentNode<AttachmentDownloadUrlQuery, AttachmentDownloadUrlQueryVariables>;
export const AuditLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AuditLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"actorUserID"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"action"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"occurredAtStart"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Time"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"occurredAtEnd"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Time"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"50"}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"0"}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"auditLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"requestID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}}},{"kind":"Argument","name":{"kind":"Name","value":"actorUserID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"actorUserID"}}},{"kind":"Argument","name":{"kind":"Name","value":"action"},"value":{"kind":"Variable","name":{"kind":"Name","value":"action"}}},{"kind":"Argument","name":{"kind":"Name","value":"occurredAtStart"},"value":{"kind":"Variable","name":{"kind":"Name","value":"occurredAtStart"}}},{"kind":"Argument","name":{"kind":"Name","value":"occurredAtEnd"},"value":{"kind":"Variable","name":{"kind":"Name","value":"occurredAtEnd"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"actorUserID"}},{"kind":"Field","name":{"kind":"Name","value":"action"}},{"kind":"Field","name":{"kind":"Name","value":"data"}},{"kind":"Field","name":{"kind":"Name","value":"occurredAt"}}]}}]}}]} as unknown as DocumentNode<AuditLogsQuery, AuditLogsQueryVariables>;
export const RequestCommentsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"RequestComments"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"comments"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"requestID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"requestID"}},{"kind":"Field","name":{"kind":"Name","value":"userID"}},{"kind":"Field","name":{"kind":"Name","value":"content"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]} as unknown as DocumentNode<RequestCommentsQuery, RequestCommentsQueryVariables>;
export const AddCommentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddComment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"content"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addComment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"requestID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}}},{"kind":"Argument","name":{"kind":"Name","value":"content"},"value":{"kind":"Variable","name":{"kind":"Name","value":"content"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"requestID"}},{"kind":"Field","name":{"kind":"Name","value":"userID"}},{"kind":"Field","name":{"kind":"Name","value":"content"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]} as unknown as DocumentNode<AddCommentMutation, AddCommentMutationVariables>;
export const DashboardSummaryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DashboardSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"dashboardSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"draftCount"}},{"kind":"Field","name":{"kind":"Name","value":"submittedCount"}},{"kind":"Field","name":{"kind":"Name","value":"approvedCount"}},{"kind":"Field","name":{"kind":"Name","value":"rejectedCount"}},{"kind":"Field","name":{"kind":"Name","value":"totalCount"}},{"kind":"Field","name":{"kind":"Name","value":"avgApprovalSeconds"}}]}}]}}]} as unknown as DocumentNode<DashboardSummaryQuery, DashboardSummaryQueryVariables>;
export const CountRequestsByStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"CountRequestsByStatus"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"countRequestsByStatus"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"count"}}]}}]}}]} as unknown as DocumentNode<CountRequestsByStatusQuery, CountRequestsByStatusQueryVariables>;
export const CountRequestsByMonthDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"CountRequestsByMonth"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"startDate"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Time"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"endDate"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Time"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"countRequestsByMonth"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"startDate"},"value":{"kind":"Variable","name":{"kind":"Name","value":"startDate"}}},{"kind":"Argument","name":{"kind":"Name","value":"endDate"},"value":{"kind":"Variable","name":{"kind":"Name","value":"endDate"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"month"}},{"kind":"Field","name":{"kind":"Name","value":"count"}}]}}]}}]} as unknown as DocumentNode<CountRequestsByMonthQuery, CountRequestsByMonthQueryVariables>;
export const AvgTimeToApprovalDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AvgTimeToApproval"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"avgTimeToApproval"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"avgSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"sampleCount"}}]}}]}}]} as unknown as DocumentNode<AvgTimeToApprovalQuery, AvgTimeToApprovalQueryVariables>;
export const MeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Me"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"me"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"userID"}},{"kind":"Field","name":{"kind":"Name","value":"orgID"}},{"kind":"Field","name":{"kind":"Name","value":"role"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}}]}}]} as unknown as DocumentNode<MeQuery, MeQueryVariables>;
export const RequestDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Request"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"request"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"orgID"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdByUserID"}},{"kind":"Field","name":{"kind":"Name","value":"decidedByUserID"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}},{"kind":"Field","name":{"kind":"Name","value":"submittedAt"}},{"kind":"Field","name":{"kind":"Name","value":"decidedAt"}},{"kind":"Field","name":{"kind":"Name","value":"steps"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"stepIndex"}},{"kind":"Field","name":{"kind":"Name","value":"label"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"assignedToUserID"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"auditTrail"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"actorUserID"}},{"kind":"Field","name":{"kind":"Name","value":"action"}},{"kind":"Field","name":{"kind":"Name","value":"data"}},{"kind":"Field","name":{"kind":"Name","value":"occurredAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"attachments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"requestID"}},{"kind":"Field","name":{"kind":"Name","value":"filename"}},{"kind":"Field","name":{"kind":"Name","value":"contentType"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"uploadedByUserID"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]}}]} as unknown as DocumentNode<RequestQuery, RequestQueryVariables>;
export const RequestsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Requests"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"50"}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"0"}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"requests"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<RequestsQuery, RequestsQueryVariables>;
export const SearchRequestsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SearchRequests"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"title"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"status"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"createdAtStart"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Time"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"createdAtEnd"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Time"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"50"}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}},"defaultValue":{"kind":"IntValue","value":"0"}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"searchRequests"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"title"},"value":{"kind":"Variable","name":{"kind":"Name","value":"title"}}},{"kind":"Argument","name":{"kind":"Name","value":"status"},"value":{"kind":"Variable","name":{"kind":"Name","value":"status"}}},{"kind":"Argument","name":{"kind":"Name","value":"createdAtStart"},"value":{"kind":"Variable","name":{"kind":"Name","value":"createdAtStart"}}},{"kind":"Argument","name":{"kind":"Name","value":"createdAtEnd"},"value":{"kind":"Variable","name":{"kind":"Name","value":"createdAtEnd"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"orgID"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdByUserID"}},{"kind":"Field","name":{"kind":"Name","value":"decidedByUserID"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}},{"kind":"Field","name":{"kind":"Name","value":"submittedAt"}},{"kind":"Field","name":{"kind":"Name","value":"decidedAt"}}]}}]}}]} as unknown as DocumentNode<SearchRequestsQuery, SearchRequestsQueryVariables>;
export const UploadAttachmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UploadAttachment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"file"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Upload"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"uploadAttachment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"requestID"},"value":{"kind":"Variable","name":{"kind":"Name","value":"requestID"}}},{"kind":"Argument","name":{"kind":"Name","value":"file"},"value":{"kind":"Variable","name":{"kind":"Name","value":"file"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"requestID"}},{"kind":"Field","name":{"kind":"Name","value":"filename"}},{"kind":"Field","name":{"kind":"Name","value":"contentType"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]} as unknown as DocumentNode<UploadAttachmentMutation, UploadAttachmentMutationVariables>;