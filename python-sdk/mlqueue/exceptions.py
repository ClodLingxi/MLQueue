"""
MLQueue 自定义异常类
"""


class MLQueueException(Exception):
    """MLQueue基础异常类"""
    pass


class ConfigurationError(MLQueueException):
    """配置错误"""
    pass


class ConnectionError(MLQueueException):
    """连接云端服务失败"""
    pass


class TaskError(MLQueueException):
    """任务相关错误"""
    pass


class QueueError(MLQueueException):
    """队列操作错误"""
    pass


class AuthenticationError(MLQueueException):
    """认证失败"""
    pass


class UploadError(MLQueueException):
    """上传失败"""
    pass
