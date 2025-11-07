"""
MLQueue - ML训练队列管理系统

一个用于管理和协调机器学习训练任务的Python库。
支持配置上传、任务队列、优先级调整、云端训练管理等功能。

支持两种架构：
- V1: 云端调度模式（适合简单任务，云端资源）
- V2: Python驱动模式（适合复杂训练，本地GPU控制）
"""

__version__ = "0.2.0"
__author__ = "MLQueue Team"

# V1 API (云端调度模式)
from .client import MLQueueClient
from .config import TrainingConfig
from .task import TrainingTask, TaskStatus
from .queue import TrainingQueue
from .trainer import MLTrainer, BatchTrainingContext, TrainingContext, train_task

# V2 API (Python驱动模式)
from .v2_client import MLQueueV2Client
from .v2_models import Group, TrainingUnit, TrainingQueue as V2TrainingQueue, QueueStatus, CreatedBy
from .exceptions import (
    MLQueueException,
    ConfigurationError,
    ConnectionError,
    TaskError,
    QueueError,
    AuthenticationError,
    UploadError
)
from .utils import generate_config_hash, validate_config, format_result

__all__ = [
    # V1 主要类
    'MLTrainer',
    'MLQueueClient',
    'TrainingQueue',
    'TrainingConfig',
    'TrainingTask',
    'BatchTrainingContext',
    'TrainingContext',

    # V2 主要类
    'MLQueueV2Client',
    'Group',
    'TrainingUnit',
    'V2TrainingQueue',

    # 枚举
    'TaskStatus',
    'QueueStatus',
    'CreatedBy',

    # 装饰器
    'train_task',

    # 异常
    'MLQueueException',
    'ConfigurationError',
    'ConnectionError',
    'TaskError',
    'QueueError',
    'AuthenticationError',
    'UploadError',

    # 工具函数
    'generate_config_hash',
    'validate_config',
    'format_result',
]
