"""
训练任务模块
"""
from typing import Optional, Dict, Any, Callable
from datetime import datetime
from enum import Enum

from .config import TrainingConfig


class TaskStatus(Enum):
    """任务状态枚举"""
    PENDING = "pending"          # 等待中
    QUEUED = "queued"            # 已入队
    RUNNING = "running"          # 运行中
    COMPLETED = "completed"      # 已完成
    FAILED = "failed"            # 失败
    CANCELLED = "cancelled"      # 已取消


class TrainingTask:
    """训练任务类"""

    def __init__(
        self,
        name: str,
        config: TrainingConfig,
        task_id: Optional[str] = None,
        priority: int = 0
    ):
        """
        初始化训练任务

        Args:
            name: 任务名称
            config: 训练配置
            task_id: 任务ID（由云端分配）
            priority: 优先级，数值越大优先级越高
        """
        self.task_id = task_id
        self.name = name
        self.config = config
        self.priority = priority
        self.status = TaskStatus.PENDING
        self.created_at = datetime.now().isoformat()
        self.started_at: Optional[str] = None
        self.completed_at: Optional[str] = None
        self.result: Optional[Dict[str, Any]] = None
        self.error_message: Optional[str] = None
        self.metadata: Dict[str, Any] = {}

    def set_status(self, status: TaskStatus) -> 'TrainingTask':
        """
        设置任务状态

        Args:
            status: 任务状态

        Returns:
            self，支持链式调用
        """
        self.status = status
        if status == TaskStatus.RUNNING and not self.started_at:
            self.started_at = datetime.now().isoformat()
        elif status in [TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED]:
            if not self.completed_at:
                self.completed_at = datetime.now().isoformat()
        return self

    def set_result(self, result: Dict[str, Any]) -> 'TrainingTask':
        """
        设置任务结果

        Args:
            result: 训练结果

        Returns:
            self，支持链式调用
        """
        self.result = result
        return self

    def set_error(self, error_message: str) -> 'TrainingTask':
        """
        设置错误信息

        Args:
            error_message: 错误消息

        Returns:
            self，支持链式调用
        """
        self.error_message = error_message
        self.status = TaskStatus.FAILED
        return self

    def to_dict(self) -> Dict[str, Any]:
        """
        转换为字典

        Returns:
            任务信息字典
        """
        return {
            'task_id': self.task_id,
            'name': self.name,
            'config': self.config.to_dict(),
            'priority': self.priority,
            'status': self.status.value,
            'created_at': self.created_at,
            'started_at': self.started_at,
            'completed_at': self.completed_at,
            'result': self.result,
            'error_message': self.error_message,
            'metadata': self.metadata
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'TrainingTask':
        """
        从字典创建任务对象

        Args:
            data: 任务数据字典

        Returns:
            TrainingTask对象
        """
        config = TrainingConfig.from_dict(data.get('config', {}))
        task = cls(
            name=data.get('name', ''),
            config=config,
            task_id=data.get('task_id'),
            priority=data.get('priority', 0)
        )
        task.status = TaskStatus(data.get('status', 'pending'))
        task.created_at = data.get('created_at', task.created_at)
        task.started_at = data.get('started_at')
        task.completed_at = data.get('completed_at')
        task.result = data.get('result')
        task.error_message = data.get('error_message')
        task.metadata = data.get('metadata', {})
        return task

    def __repr__(self) -> str:
        return f"TrainingTask(id={self.task_id}, name={self.name}, status={self.status.value})"
