"""
训练配置管理模块
"""
from typing import Any, Dict, Optional
import json
from datetime import datetime


class TrainingConfig:
    """训练配置类"""

    def __init__(self, config_dict: Optional[Dict[str, Any]] = None, **kwargs):
        """
        初始化训练配置

        Args:
            config_dict: 配置字典
            **kwargs: 额外的配置参数
        """
        self.config: Dict[str, Any] = config_dict or {}
        self.config.update(kwargs)
        self.created_at = datetime.now().isoformat()
        self.config_id: Optional[str] = None

    def set(self, key: str, value: Any) -> 'TrainingConfig':
        """
        设置配置项

        Args:
            key: 配置键
            value: 配置值

        Returns:
            self，支持链式调用
        """
        self.config[key] = value
        return self

    def get(self, key: str, default: Any = None) -> Any:
        """
        获取配置项

        Args:
            key: 配置键
            default: 默认值

        Returns:
            配置值
        """
        return self.config.get(key, default)

    def update(self, config_dict: Dict[str, Any]) -> 'TrainingConfig':
        """
        批量更新配置

        Args:
            config_dict: 配置字典

        Returns:
            self，支持链式调用
        """
        self.config.update(config_dict)
        return self

    def to_dict(self) -> Dict[str, Any]:
        """
        转换为字典

        Returns:
            配置字典
        """
        return {
            'config_id': self.config_id,
            'config': self.config,
            'created_at': self.created_at
        }

    def to_json(self) -> str:
        """
        转换为JSON字符串

        Returns:
            JSON字符串
        """
        return json.dumps(self.to_dict(), ensure_ascii=False, indent=2)

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'TrainingConfig':
        """
        从字典创建配置对象

        Args:
            data: 包含配置信息的字典

        Returns:
            TrainingConfig对象
        """
        config = cls(data.get('config', {}))
        config.config_id = data.get('config_id')
        config.created_at = data.get('created_at', datetime.now().isoformat())
        return config

    def __repr__(self) -> str:
        return f"TrainingConfig(id={self.config_id}, config={self.config})"
