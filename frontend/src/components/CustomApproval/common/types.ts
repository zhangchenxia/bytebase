export type CustomApprovalNodeType = 
  | "APPROVAL" 
  | "CONDITION" 
  | "NOTIFICATION" 
  | "ACTION"; 

export interface CustomApprovalNode {
  id: string;
  name: string;
  description?: string;
  type: CustomApprovalNodeType;
  config?: Record<string, any>;
  createdAt: number;
  updatedAt: number;
}

export interface CustomApprovalNodeConfig {
  // For APPROVAL node
  approvers?: string[];
  approvalType?: "AND" | "OR";
  
  // For CONDITION node
  condition?: string; // CEL expression
  
  // For NOTIFICATION node
  notificationType?: string;
  recipients?: string[];
  
  // For ACTION node
  actionType?: string;
  actionConfig?: Record<string, any>;
}

export interface CreateCustomApprovalNodeRequest {
  name: string;
  description?: string;
  type: CustomApprovalNodeType;
  config?: CustomApprovalNodeConfig;
}

export interface UpdateCustomApprovalNodeRequest {
  id: string;
  name?: string;
  description?: string;
  type?: CustomApprovalNodeType;
  config?: CustomApprovalNodeConfig;
}
