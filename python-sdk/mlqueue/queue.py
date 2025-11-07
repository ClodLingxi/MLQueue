"""
训练队列管理模块
"""
from typing import List, Optional, Dict, Any
from .task import TrainingTask, TaskStatus
from .config import TrainingConfig
from .client import MLQueueClient
from .exceptions import QueueError


class TrainingQueue:
    """训练队列管理器"""

    def __init__(self, client: MLQueueClient, queue_name: str = "default"):
        """
        初始化训练队列

        Args:
            client: MLQueue客户端
            queue_name: 队列名称
        """
        self.client = client
        self.queue_name = queue_name
        self.tasks: List[TrainingTask] = []

    def add_task(
        self,
        name: str,
        config: TrainingConfig,
        priority: int = 0,
        upload: bool = False
    ) -> TrainingTask:
        """
        添加训练任务到队列

        Args:
            name: 任务名称
            config: 训练配置
            priority: 优先级
            upload: 是否立即上传到云端

        Returns:
            TrainingTask对象

        Raises:
            QueueError: 添加失败
        """
        task = TrainingTask(name=name, config=config, priority=priority)
        self.tasks.append(task)

        if upload:
            try:
                self.client.create_task(task)
                task.set_status(TaskStatus.QUEUED)
            except Exception as e:
                raise QueueError(f"上传任务失败: {str(e)}")

        return task

    def add_tasks(
        self,
        task_configs: List[Dict[str, Any]],
        upload: bool = False
    ) -> List[TrainingTask]:
        """
        批量添加任务

        Args:
            task_configs: 任务配置列表，每个元素包含name, config, priority
            upload: 是否立即上传到云端

        Returns:
            TrainingTask对象列表

        Raises:
            QueueError: 添加失败
        """
        tasks = []
        for task_config in task_configs:
            task = TrainingTask(
                name=task_config.get('name', 'Unnamed'),
                config=task_config.get('config'),
                priority=task_config.get('priority', 0)
            )
            tasks.append(task)
            self.tasks.append(task)

        if upload:
            try:
                self.client.batch_create_tasks(tasks)
                for task in tasks:
                    task.set_status(TaskStatus.QUEUED)
            except Exception as e:
                raise QueueError(f"批量上传任务失败: {str(e)}")

        return tasks

    def upload_all(self) -> List[str]:
        """
        上传所有未上传的任务到云端

        Returns:
            任务ID列表

        Raises:
            QueueError: 上传失败
        """
        pending_tasks = [task for task in self.tasks if not task.task_id]
        if not pending_tasks:
            return []

        try:
            task_ids = self.client.batch_create_tasks(pending_tasks)
            for task in pending_tasks:
                task.set_status(TaskStatus.QUEUED)
            return task_ids
        except Exception as e:
            raise QueueError(f"上传队列失败: {str(e)}")

    def get_status(self) -> Dict[str, Any]:
        """
        获取队列状态

        Returns:
            队列状态信息
        """
        try:
            return self.client.get_queue_status()
        except Exception as e:
            raise QueueError(f"获取队列状态失败: {str(e)}")

    def list_tasks(self, status: Optional[TaskStatus] = None) -> List[TrainingTask]:
        """
        列出队列中的任务

        Args:
            status: 状态过滤

        Returns:
            任务列表
        """
        try:
            return self.client.list_tasks(status=status)
        except Exception as e:
            raise QueueError(f"获取任务列表失败: {str(e)}")

    def update_priority(self, task_id: str, priority: int) -> bool:
        """
        更新任务优先级

        Args:
            task_id: 任务ID
            priority: 新的优先级

        Returns:
            是否成功
        """
        try:
            return self.client.update_task_priority(task_id, priority)
        except Exception as e:
            raise QueueError(f"更新优先级失败: {str(e)}")

    def cancel_task(self, task_id: str) -> bool:
        """
        取消任务

        Args:
            task_id: 任务ID

        Returns:
            是否成功
        """
        try:
            return self.client.cancel_task(task_id)
        except Exception as e:
            raise QueueError(f"取消任务失败: {str(e)}")

    def reorder(self, task_ids: List[str]) -> bool:
        """
        重新排列队列

        Args:
            task_ids: 任务ID列表（新的顺序）

        Returns:
            是否成功
        """
        try:
            return self.client.reorder_queue(task_ids)
        except Exception as e:
            raise QueueError(f"重排队列失败: {str(e)}")

    def clear(self):
        """清空本地任务列表"""
        self.tasks.clear()

    def __len__(self) -> int:
        return len(self.tasks)

    def __repr__(self) -> str:
        return f"TrainingQueue(name={self.queue_name}, tasks={len(self.tasks)})"
