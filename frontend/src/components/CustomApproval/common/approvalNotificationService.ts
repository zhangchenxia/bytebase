import { sendApprovalRequestNotification, sendApprovalResultNotification, sendApprovalCompletedNotification, sendApprovalCancelledNotification } from "./notification"; 
import type { ApprovalRequest, ApprovalResult, ApprovalFlowEvent } from "./types"; 

/**
 * 审批流程通知服务
 * 负责在审批流程的关键节点发送通知
 */
export class ApprovalNotificationService {
  /**
   * 当有新的审批请求时发送通知
   * @param request 审批请求信息
   */
  static sendNewApprovalRequest(request: ApprovalRequest): void {
    // 向审批人发送通知
    request.approvers.forEach(approver => {
      sendApprovalRequestNotification(
        request.requester.name,
        request.title
      );
    });
  }

  /**
   * 当审批结果更新时发送通知
   * @param result 审批结果信息
   */
  static sendApprovalResult(result: ApprovalResult): void {
    // 向请求人发送通知
    sendApprovalResultNotification(
      result.requestTitle,
      result.isApproved,
      result.approver.name,
      result.comment
    );

    // 如果有下一级审批人，向他们发送通知
    if (result.nextApprovers && result.nextApprovers.length > 0 && result.isApproved) {
      result.nextApprovers.forEach(approver => {
        sendApprovalRequestNotification(
          result.requester.name,
          result.requestTitle
        );
      });
    }
  }

  /**
   * 当审批流程完成时发送通知
   * @param requestTitle 审批请求标题
   * @param requesterName 请求人姓名
   * @param isApproved 是否通过
   */
  static sendApprovalCompleted(requestTitle: string, requesterName: string, isApproved: boolean): void {
    // 向请求人发送完成通知
    if (isApproved) {
      sendApprovalCompletedNotification(requestTitle, requesterName);
    } else {
      // 如果被拒绝，已经在审批结果通知中处理了
    }
  }

  /**
   * 当审批流程取消时发送通知
   * @param requestTitle 审批请求标题
   * @param requesterName 请求人姓名
   */
  static sendApprovalCancelled(requestTitle: string, requesterName: string): void {
    // 向请求人和所有审批人发送取消通知
    sendApprovalCancelledNotification(requestTitle, requesterName);
  }

  /**
   * 处理审批流程事件并发送相应通知
   * @param event 审批流程事件
   */
  static handleApprovalEvent(event: ApprovalFlowEvent): void {
    switch (event.type) {
      case "approval-requested":
        if (event.data.request) {
          this.sendNewApprovalRequest(event.data.request);
        }
        break;

      case "approval-completed":
        if (event.data.result) {
          this.sendApprovalResult(event.data.result);
        }
        break;

      case "flow-completed":
        if (event.data) {
          this.sendApprovalCompleted(
            event.data.requestTitle,
            event.data.requesterName,
            event.data.isApproved
          );
        }
        break;

      case "flow-cancelled":
        if (event.data) {
          this.sendApprovalCancelled(
            event.data.requestTitle,
            event.data.requesterName
          );
        }
        break;

      default:
        console.warn(`Unknown approval event type: ${event.type}`);
        break;
    }
  }
}

// 审批相关类型定义
export interface User {
  id: string;
  name: string;
  email: string;
}

export interface ApprovalRequest {
  id: string;
  title: string;
  description?: string;
  requester: User;
  approvers: User[];
  createdAt: number;
}

export interface ApprovalResult {
  id: string;
  requestId: string;
  requestTitle: string;
  approver: User;
  requester: User;
  isApproved: boolean;
  comment?: string;
  completedAt: number;
  nextApprovers?: User[];
}

export interface ApprovalFlowEvent {
  type: "approval-requested" | "approval-completed" | "flow-completed" | "flow-cancelled";
  data: {
    request?: ApprovalRequest;
    result?: ApprovalResult;
    requestTitle?: string;
    requesterName?: string;
    isApproved?: boolean;
  };
}