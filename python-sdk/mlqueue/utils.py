"""
工具函数模块
"""
from typing import Any, Dict
import json
import hashlib


def generate_config_hash(config: Dict[str, Any]) -> str:
    """
    生成配置哈希值（用于去重）

    Args:
        config: 配置字典

    Returns:
        配置的MD5哈希值
    """
    config_str = json.dumps(config, sort_keys=True)
    return hashlib.md5(config_str.encode()).hexdigest()


def validate_config(config: Dict[str, Any], required_keys: list[str]) -> bool:
    """
    验证配置是否包含必需的键

    Args:
        config: 配置字典
        required_keys: 必需的键列表

    Returns:
        是否有效
    """
    return all(key in config for key in required_keys)


def format_result(result: Dict[str, Any], precision: int = 4) -> Dict[str, Any]:
    """
    格式化训练结果（保留指定精度）

    Args:
        result: 结果字典
        precision: 小数精度

    Returns:
        格式化后的结果字典
    """
    formatted = {}
    for key, value in result.items():
        if isinstance(value, float):
            formatted[key] = round(value, precision)
        elif isinstance(value, dict):
            formatted[key] = format_result(value, precision)
        else:
            formatted[key] = value
    return formatted
