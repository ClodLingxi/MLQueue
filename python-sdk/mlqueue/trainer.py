"""
训练器包装模块
提供便捷的装饰器和上下文管理器
"""
from typing import Callable, Optional, Dict, Any, List, Iterator
from functools import wraps
import traceback
import time

from .client import MLQueueClient
from .queue import TrainingQueue
from .config import TrainingConfig
from .task import TrainingTask, TaskStatus


class MLTrainer:
    """ML训练器包装类"""

    def __init__(
        self,
        api_url: str,
        api_key: Optional[str] = None,
        queue_name: str = "default",
        auto_upload: bool = True
    ):
        """
        初始化训练器

        Args:
            api_url: 云端API地址
            api_key: API密钥
            queue_name: 队列名称
            auto_upload: 是否自动上传配置和结果
        """
        self.client = MLQueueClient(api_url=api_url, api_key=api_key)
        self.queue = TrainingQueue(client=self.client, queue_name=queue_name)
        self.auto_upload = auto_upload
        self.current_task: Optional[TrainingTask] = None

    def start_training(
        self,
        batch_name: str,
        configs: List[Dict[str, Any]],
        train_fn: Optional[Callable] = None,
        base_priority: int = 0,
        poll_interval: int = 5
    ) -> 'BatchTrainingContext':
        """
        开始批量训练（返回批量训练上下文管理器）

        Args:
            batch_name: 整体训练批次名称
            configs: 训练配置列表，每个配置可以是Dict或包含name的Dict
            train_fn: 可选的训练函数，如果提供则自动执行所有配置
            base_priority: 基础优先级
            poll_interval: 从云端轮询队列的间隔时间（秒）

        Returns:
            BatchTrainingContext批量训练上下文管理器

        Example:
            # 方式1: 手动迭代
            configs = [{"hidden_size": 64}, {"hidden_size": 128}]
            with trainer.start_training("实验批次", configs) as batch:
                for config in batch:
                    result = train_model(**config)
                    batch.log_result(result)

            # 方式2: 自动执行
            def train_model(config):
                # 训练逻辑
                return {"loss": 0.1}

            with trainer.start_training("实验批次", configs, train_fn=train_model):
                pass  # 自动执行所有配置
        """
        # 准备任务列表
        task_configs = []
        for i, cfg in enumerate(configs):
            if isinstance(cfg, dict):
                # 检查是否包含name字段
                if 'name' in cfg and 'config' in cfg:
                    task_configs.append({
                        'name': cfg['name'],
                        'config': TrainingConfig(cfg['config']),
                        'priority': cfg.get('priority', base_priority)
                    })
                else:
                    # 直接作为配置
                    task_configs.append({
                        'name': f"{batch_name}_{i+1}",
                        'config': TrainingConfig(cfg),
                        'priority': base_priority
                    })

        # 批量创建任务并上传到云端
        tasks = []
        if self.auto_upload:
            tasks = self.queue.add_tasks(task_configs, upload=True)
        else:
            for tc in task_configs:
                task = TrainingTask(
                    name=tc['name'],
                    config=tc['config'],
                    priority=tc['priority']
                )
                tasks.append(task)

        return BatchTrainingContext(
            trainer=self,
            batch_name=batch_name,
            tasks=tasks,
            train_fn=train_fn,
            poll_interval=poll_interval
        )

    def add_config(
        self,
        name: str,
        config: Dict[str, Any],
        priority: int = 0
    ) -> TrainingTask:
        """
        添加训练配置（不立即执行）

        Args:
            name: 训练名称
            config: 训练配置
            priority: 优先级

        Returns:
            TrainingTask对象
        """
        training_config = TrainingConfig(config)
        return self.queue.add_task(
            name=name,
            config=training_config,
            priority=priority,
            upload=self.auto_upload
        )

    def add_configs(
        self,
        configs: list[Dict[str, Any]]
    ) -> list[TrainingTask]:
        """
        批量添加训练配置

        Args:
            configs: 配置列表，每个元素包含name, config, priority等字段

        Returns:
            TrainingTask对象列表
        """
        task_configs = []
        for cfg in configs:
            training_config = TrainingConfig(cfg.get('config', {}))
            task_configs.append({
                'name': cfg.get('name', 'Unnamed'),
                'config': training_config,
                'priority': cfg.get('priority', 0)
            })

        return self.queue.add_tasks(task_configs, upload=self.auto_upload)

    def upload_result(self, task_id: str, result: Dict[str, Any]) -> bool:
        """
        上传训练结果

        Args:
            task_id: 任务ID
            result: 训练结果

        Returns:
            是否成功
        """
        return self.client.upload_result(task_id, result)

    def get_queue_status(self) -> Dict[str, Any]:
        """获取队列状态"""
        return self.queue.get_status()

    def list_tasks(self, status=None) -> list[TrainingTask]:
        """列出任务"""
        return self.queue.list_tasks(status=status)


class BatchTrainingContext:
    """批量训练上下文管理器"""

    def __init__(
        self,
        trainer: MLTrainer,
        batch_name: str,
        tasks: List[TrainingTask],
        train_fn: Optional[Callable] = None,
        poll_interval: int = 5
    ):
        """
        初始化批量训练上下文

        Args:
            trainer: MLTrainer对象
            batch_name: 批次名称
            tasks: 任务列表
            train_fn: 训练函数
            poll_interval: 轮询间隔（秒）
        """
        self.trainer = trainer
        self.batch_name = batch_name
        self.tasks = tasks
        self.train_fn = train_fn
        self.poll_interval = poll_interval
        self.current_task: Optional[TrainingTask] = None
        self.current_result: Optional[Dict[str, Any]] = None
        self.completed_tasks: List[str] = []

    def __enter__(self) -> 'BatchTrainingContext':
        """进入上下文"""
        print(f"开始批量训练: {self.batch_name}")
        print(f"共 {len(self.tasks)} 个训练任务已上传到云端")

        # 如果提供了训练函数，自动执行所有配置
        if self.train_fn:
            self._auto_execute()

        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """退出上下文"""
        if exc_type is not None:
            # 发生异常
            error_msg = ''.join(traceback.format_exception(exc_type, exc_val, exc_tb))
            print(f"批量训练发生错误: {error_msg}")
            if self.current_task and self.current_task.task_id:
                self.trainer.client.upload_result(
                    self.current_task.task_id,
                    {'error': error_msg, 'status': 'failed'}
                )
            return False

        print(f"批量训练完成: 已完成 {len(self.completed_tasks)} 个任务")
        return True

    def _fetch_queue_from_cloud(self) -> List[TrainingTask]:
        """
        从云端获取最新的队列顺序

        Returns:
            按云端顺序排列的待执行任务列表
        """
        try:
            # 获取所有排队中的任务
            queued_tasks = self.trainer.client.list_tasks(status=TaskStatus.QUEUED)

            # 过滤出属于当前批次的任务
            batch_task_ids = {task.task_id for task in self.tasks if task.task_id}
            batch_queued_tasks = [
                task for task in queued_tasks
                if task.task_id in batch_task_ids
            ]

            return batch_queued_tasks

        except Exception as e:
            print(f"从云端获取队列失败: {e}")
            # 降级方案：返回本地任务列表
            return [task for task in self.tasks if task.task_id not in self.completed_tasks]

    def __iter__(self) -> Iterator[Dict[str, Any]]:
        """
        迭代器，每次迭代前从云端获取最新队列

        Yields:
            训练配置字典
        """
        while len(self.completed_tasks) < len(self.tasks):
            # 从云端获取最新队列
            print(f"\n从云端获取最新训练队列...")
            remaining_tasks = self._fetch_queue_from_cloud()

            if not remaining_tasks:
                print("所有任务已完成或被取消")
                break

            # 取第一个任务（优先级最高的）
            self.current_task = remaining_tasks[0]
            print(f"执行任务: {self.current_task.name} (ID: {self.current_task.task_id})")
            print(f"配置: {self.current_task.config.config}")

            # 标记为运行中
            if self.current_task.task_id and self.trainer.auto_upload:
                try:
                    # 这里可以添加一个更新任务状态为RUNNING的API调用
                    pass
                except Exception as e:
                    print(f"更新任务状态失败: {e}")

            # 返回配置供用户训练
            yield self.current_task.config.config

            # 等待一段时间再获取下一个任务（避免频繁请求）
            if len(self.completed_tasks) < len(self.tasks):
                time.sleep(self.poll_interval)

    def log_result(self, result: Dict[str, Any]):
        """
        记录当前任务的训练结果

        Args:
            result: 训练结果字典
        """
        if not self.current_task:
            print("警告: 没有正在执行的任务")
            return

        self.current_result = result
        self.current_task.set_result(result)
        self.current_task.set_status(TaskStatus.COMPLETED)

        # 上传结果到云端
        if self.current_task.task_id and self.trainer.auto_upload:
            try:
                self.trainer.upload_result(self.current_task.task_id, result)
                print(f"结果已上传: {result}")
            except Exception as e:
                print(f"上传结果失败: {e}")

        # 标记为已完成
        if self.current_task.task_id:
            self.completed_tasks.append(self.current_task.task_id)

    def _auto_execute(self):
        """自动执行所有训练配置"""
        if not self.train_fn:
            return

        print("开始自动执行训练...")
        for config in self:
            try:
                print(f"\n训练配置: {config}")
                result = self.train_fn(config)
                if isinstance(result, dict):
                    self.log_result(result)
                else:
                    print(f"警告: 训练函数返回值不是字典: {result}")
            except Exception as e:
                error_msg = f"训练失败: {str(e)}\n{traceback.format_exc()}"
                print(error_msg)
                if self.current_task and self.current_task.task_id:
                    self.trainer.upload_result(
                        self.current_task.task_id,
                        {'error': error_msg, 'status': 'failed'}
                    )

    def get_progress(self) -> Dict[str, Any]:
        """
        获取训练进度

        Returns:
            进度信息字典
        """
        return {
            'batch_name': self.batch_name,
            'total': len(self.tasks),
            'completed': len(self.completed_tasks),
            'remaining': len(self.tasks) - len(self.completed_tasks),
            'progress': len(self.completed_tasks) / len(self.tasks) if self.tasks else 0
        }


class TrainingContext:
    """单次训练上下文管理器（已废弃，保留用于向后兼容）"""

    def __init__(self, trainer: MLTrainer, task: TrainingTask):
        """
        初始化上下文

        Args:
            trainer: MLTrainer对象
            task: TrainingTask对象
        """
        self.trainer = trainer
        self.task = task
        self.result: Optional[Dict[str, Any]] = None

    def __enter__(self) -> 'TrainingContext':
        """进入上下文"""
        self.task.set_status(TaskStatus.RUNNING)
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """退出上下文"""
        if exc_type is not None:
            # 发生异常
            error_msg = ''.join(traceback.format_exception(exc_type, exc_val, exc_tb))
            self.task.set_error(error_msg)
            if self.trainer.auto_upload and self.task.task_id:
                self.trainer.upload_result(
                    self.task.task_id,
                    {'error': error_msg, 'status': 'failed'}
                )
            return False
        else:
            # 正常完成
            self.task.set_status(TaskStatus.COMPLETED)
            if self.result:
                self.task.set_result(self.result)
                if self.trainer.auto_upload and self.task.task_id:
                    self.trainer.upload_result(self.task.task_id, self.result)
        return True

    def log_result(self, result: Dict[str, Any]):
        """
        记录训练结果

        Args:
            result: 训练结果字典
        """
        self.result = result

    def get_config(self) -> Dict[str, Any]:
        """
        获取当前训练配置

        Returns:
            配置字典
        """
        return self.task.config.config


def train_task(
    trainer: MLTrainer,
    name: str,
    priority: int = 0,
    auto_log: bool = True
) -> Callable:
    """
    训练任务装饰器

    Args:
        trainer: MLTrainer对象
        name: 任务名称
        priority: 优先级
        auto_log: 是否自动记录返回值作为结果

    Returns:
        装饰器函数

    Example:
        @train_task(trainer, "train_model", priority=1)
        def train(config):
            # 训练代码
            return {"loss": 0.1, "accuracy": 0.95}
    """
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(config: Dict[str, Any], **kwargs):
            with trainer.start_training(name, config, priority) as ctx:
                result = func(config, **kwargs)
                if auto_log and isinstance(result, dict):
                    ctx.log_result(result)
                return result
        return wrapper
    return decorator
