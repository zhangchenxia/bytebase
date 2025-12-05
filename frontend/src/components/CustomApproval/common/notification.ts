import { pushNotification } from "@/store";
import type { BBNotificationStyle } from "@/types";

const APPROVAL_NOTIFICATION_MODULE = "custom-approval";

export type ApprovalNotificationType = 
  | "approval-requested"
  | "approval-approved"
  | "approval-rejected"
  | "approval-cancelled"
  | "approval-completed";

interface ApprovalNotificationConfig {
  type: ApprovalNotificationType;
  title: string;
  description?: string;
  style: BBNotificationStyle;
  link?: string;
  linkTitle?: string;
}

const notificationConfigs: Record<ApprovalNotificationType, ApprovalNotificationConfig> = {
  "approval-requested": {
    type: "approval-requested",
    title: "新的审批请求",
    description: "您有一个新的审批请求需要处理。",
    style: "INFO",
    link: "/approval",
    linkTitle: "查看审批"
  },
  "approval-approved": {
    type: "approval-approved",
    title: "审批已通过",
    description: "您的审批请求已通过。",
    style: "SUCCESS",
    link: "/approval",
    linkTitle: "查看详情"
  },
  "approval-rejected": {
    type: "approval-rejected",
    title: "审批已拒绝",
    description: "您的审批请求已被拒绝。",
    style: "WARN",
    link: "/approval",
    linkTitle: "查看详情"
  },
  "approval-cancelled": {
    type: "approval-cancelled",
    title: "审批已取消",
    description: "您的审批请求已取消。",
    style: "INFO",
    link: "/approval",
    linkTitle: "查看详情"
  },
  "approval-completed": {
    type: "approval-completed",
    title: "审批已完成",
    description: "审批流程已完成。",
    style: "SUCCESS",
    link: "/approval",
    linkTitle: "查看详情"
  }
};

export const sendApprovalNotification = (
  type: ApprovalNotificationType,
  customTitle?: string,
  customDescription?: string
) => {
  const config = notificationConfigs[type];
  
  if (!config) {
    console.warn(`Unknown approval notification type: ${type}`);
    return;
  }
  
  pushNotification({
    module: APPROVAL_NOTIFICATION_MODULE,
    style: config.style,
    title: customTitle || config.title,
    description: customDescription || config.description,
    link: config.link,
    linkTitle: config.linkTitle
  });
};

export const sendApprovalRequestNotification = (
  approverName: string,
  requestTitle: string
) => {
  sendApprovalNotification(
    "approval-requested",
    `新的审批请求需要您的处理`,
    `审批标题: ${requestTitle}\n请求人: ${approverName}`
  );
};

export const sendApprovalResultNotification = (
  requestTitle: string,
  isApproved: boolean,
  approverName: string,
  comment?: string
) => {
  const type = isApproved ? "approval-approved" : "approval-rejected";
  const result = isApproved ? "通过" : "拒绝";
  const commentText = comment ? `\n审批意见: ${comment}` : "";
  
  sendApprovalNotification(
    type,
    `审批${result}`,
    `审批标题: ${requestTitle}\n审批人: ${approverName}${commentText}`
  );
};

export const sendApprovalCompletedNotification = (
  requestTitle: string,
  requesterName: string
) => {
  sendApprovalNotification(
    "approval-completed",
    `审批流程已完成`,
    `审批标题: ${requestTitle}\n请求人: ${requesterName}`
  );
};

export const sendApprovalCancelledNotification = (
  requestTitle: string,
  requesterName: string
) => {
  sendApprovalNotification(
    "approval-cancelled",
    `审批已取消`,
    `审批标题: ${requestTitle}\n请求人: ${requesterName}`
  );
};