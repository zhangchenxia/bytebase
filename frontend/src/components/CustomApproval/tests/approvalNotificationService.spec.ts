import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ApprovalNotificationService } from '../common/approvalNotificationService';
import { pushNotification } from '@/store/modules/notification';

// Mock the pushNotification function
vi.mock('@/store/modules/notification', () => ({
  pushNotification: vi.fn(),
}));

describe('ApprovalNotificationService', () => {
  beforeEach(() => {
    // Reset all mocks before each test
    vi.clearAllMocks();
  });

  it('should send new approval request notification', () => {
    // Test data
    const user = {
      id: 'user1',
      name: 'User One',
      email: 'user1@example.com',
    };

    const approvalRequest = {
      id: 'req1',
      title: 'Test Approval Request',
      description: 'This is a test approval request',
      requester: user,
      createdAt: new Date(),
    };

    // Call the service method
    ApprovalNotificationService.sendNewApprovalRequest(approvalRequest, [user]);

    // Check if pushNotification was called with the correct data
    expect(pushNotification).toHaveBeenCalled();
    expect(pushNotification).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '新的审批请求',
        content: expect.stringContaining('Test Approval Request'),
        style: 'info',
      })
    );
  });

  it('should send approval result notification', () => {
    // Test data
    const user = {
      id: 'user1',
      name: 'User One',
      email: 'user1@example.com',
    };

    const approvalRequest = {
      id: 'req1',
      title: 'Test Approval Request',
      description: 'This is a test approval request',
      requester: user,
      createdAt: new Date(),
    };

    const approvalResult = {
      id: 'res1',
      requestId: 'req1',
      approver: user,
      approved: true,
      comment: 'Approved',
      approvedAt: new Date(),
    };

    // Call the service method
    ApprovalNotificationService.sendApprovalResult(approvalRequest, approvalResult, [user]);

    // Check if pushNotification was called with the correct data
    expect(pushNotification).toHaveBeenCalled();
    expect(pushNotification).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '审批结果',
        content: expect.stringContaining('Test Approval Request'),
        content: expect.stringContaining('已通过'),
        style: 'success',
      })
    );
  });

  it('should send rejected approval result notification', () => {
    // Test data
    const user = {
      id: 'user1',
      name: 'User One',
      email: 'user1@example.com',
    };

    const approvalRequest = {
      id: 'req1',
      title: 'Test Approval Request',
      description: 'This is a test approval request',
      requester: user,
      createdAt: new Date(),
    };

    const approvalResult = {
      id: 'res1',
      requestId: 'req1',
      approver: user,
      approved: false,
      comment: 'Rejected',
      approvedAt: new Date(),
    };

    // Call the service method
    ApprovalNotificationService.sendApprovalResult(approvalRequest, approvalResult, [user]);

    // Check if pushNotification was called with the correct data
    expect(pushNotification).toHaveBeenCalled();
    expect(pushNotification).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '审批结果',
        content: expect.stringContaining('Test Approval Request'),
        content: expect.stringContaining('已拒绝'),
        style: 'error',
      })
    );
  });

  it('should send approval flow completed notification', () => {
    // Test data
    const user = {
      id: 'user1',
      name: 'User One',
      email: 'user1@example.com',
    };

    const approvalRequest = {
      id: 'req1',
      title: 'Test Approval Request',
      description: 'This is a test approval request',
      requester: user,
      createdAt: new Date(),
    };

    // Call the service method
    ApprovalNotificationService.sendApprovalFlowCompleted(approvalRequest, true, [user]);

    // Check if pushNotification was called with the correct data
    expect(pushNotification).toHaveBeenCalled();
    expect(pushNotification).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '审批流程完成',
        content: expect.stringContaining('Test Approval Request'),
        content: expect.stringContaining('已通过'),
        style: 'success',
      })
    );
  });

  it('should send approval flow failed notification', () => {
    // Test data
    const user = {
      id: 'user1',
      name: 'User One',
      email: 'user1@example.com',
    };

    const approvalRequest = {
      id: 'req1',
      title: 'Test Approval Request',
      description: 'This is a test approval request',
      requester: user,
      createdAt: new Date(),
    };

    // Call the service method
    ApprovalNotificationService.sendApprovalFlowCompleted(approvalRequest, false, [user]);

    // Check if pushNotification was called with the correct data
    expect(pushNotification).toHaveBeenCalled();
    expect(pushNotification).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '审批流程完成',
        content: expect.stringContaining('Test Approval Request'),
        content: expect.stringContaining('未通过'),
        style: 'error',
      })
    );
  });

  it('should send node reminder notification', () => {
    // Test data
    const user = {
      id: 'user1',
      name: 'User One',
      email: 'user1@example.com',
    };

    const approvalRequest = {
      id: 'req1',
      title: 'Test Approval Request',
      description: 'This is a test approval request',
      requester: user,
      createdAt: new Date(),
    };

    const node = {
      id: 'node1',
      name: 'Test Node',
      type: 'APPROVAL',
    };

    // Call the service method
    ApprovalNotificationService.sendNodeReminder(approvalRequest, node, [user]);

    // Check if pushNotification was called with the correct data
    expect(pushNotification).toHaveBeenCalled();
    expect(pushNotification).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '审批节点提醒',
        content: expect.stringContaining('Test Approval Request'),
        content: expect.stringContaining('Test Node'),
        style: 'warning',
      })
    );
  });
});