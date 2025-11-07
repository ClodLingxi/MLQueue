"""
云端API客户端模块
"""
from typing import List, Dict, Any, Optional
import requests
import json

from .task import TrainingTask, TaskStatus
from .config import TrainingConfig
from .exceptions import (
    ConnectionError,
    AuthenticationError,
    UploadError,
    TaskError
)


class MLQueueClient:
    """ML训练队列云端客户端"""

    def __init__(
        self,
        api_url: str,
        api_key: Optional[str] = None,
        timeout: int = 30
    ):
        """
        初始化客户端

        Args:
            api_url: 云端API基础URL
            api_key: API密钥（用于身份验证）
            timeout: 请求超时时间（秒）
        """
        self.api_url = api_url.rstrip('/')
        self.api_key = api_key
        self.timeout = timeout
        self.session = requests.Session()
        if api_key:
            self.session.headers.update({
                'Authorization': f'Bearer {api_key}',
                'Content-Type': 'application/json'
            })

    def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        发送HTTP请求

        Args:
            method: HTTP方法
            endpoint: API端点
            data: 请求数据
            params: URL参数

        Returns:
            响应数据

        Raises:
            ConnectionError: 连接失败
            AuthenticationError: 认证失败
        """
        url = f"{self.api_url}/{endpoint.lstrip('/')}"
        try:
            response = self.session.request(
                method=method,
                url=url,
                json=data,
                params=params,
                timeout=self.timeout
            )

            if response.status_code == 401:
                raise AuthenticationError("认证失败，请检查API密钥")
            elif response.status_code == 403:
                raise AuthenticationError("权限不足")
            elif response.status_code >= 400:
                error_msg = response.json().get('error', response.text)
                raise ConnectionError(f"请求失败: {error_msg}")

            return response.json()

        except requests.exceptions.Timeout:
            raise ConnectionError(f"请求超时（{self.timeout}秒）")
        except requests.exceptions.ConnectionError as e:
            raise ConnectionError(f"无法连接到云端服务: {str(e)}")
        except json.JSONDecodeError:
            raise ConnectionError("无效的响应格式")

    def create_task(self, task: TrainingTask) -> str:
        """
        创建训练任务

        Args:
            task: 训练任务对象

        Returns:
            任务ID

        Raises:
            UploadError: 上传失败
        """
        try:
            response = self._request(
                method='POST',
                endpoint='/tasks',
                data=task.to_dict()
            )
            task_id = response.get('task_id')
            if not task_id:
                raise UploadError("创建任务失败：未返回任务ID")
            task.task_id = task_id
            return task_id
        except ConnectionError as e:
            raise UploadError(f"上传任务失败: {str(e)}")

    def batch_create_tasks(self, tasks: List[TrainingTask]) -> List[str]:
        """
        批量创建训练任务

        Args:
            tasks: 训练任务列表

        Returns:
            任务ID列表

        Raises:
            UploadError: 上传失败
        """
        try:
            response = self._request(
                method='POST',
                endpoint='/tasks/batch',
                data={'tasks': [task.to_dict() for task in tasks]}
            )
            task_ids = response.get('task_ids', [])
            for task, task_id in zip(tasks, task_ids):
                task.task_id = task_id
            return task_ids
        except ConnectionError as e:
            raise UploadError(f"批量上传任务失败: {str(e)}")

    def get_task(self, task_id: str) -> TrainingTask:
        """
        获取任务信息

        Args:
            task_id: 任务ID

        Returns:
            训练任务对象

        Raises:
            TaskError: 任务不存在或获取失败
        """
        try:
            response = self._request(
                method='GET',
                endpoint=f'/tasks/{task_id}'
            )
            return TrainingTask.from_dict(response)
        except ConnectionError as e:
            raise TaskError(f"获取任务失败: {str(e)}")

    def list_tasks(
        self,
        status: Optional[TaskStatus] = None,
        limit: int = 100,
        offset: int = 0
    ) -> List[TrainingTask]:
        """
        列出任务

        Args:
            status: 任务状态过滤
            limit: 返回数量限制
            offset: 偏移量

        Returns:
            训练任务列表
        """
        params = {'limit': limit, 'offset': offset}
        if status:
            params['status'] = status.value

        response = self._request(
            method='GET',
            endpoint='/tasks',
            params=params
        )
        return [TrainingTask.from_dict(task_data) for task_data in response.get('tasks', [])]

    def update_task_priority(self, task_id: str, priority: int) -> bool:
        """
        更新任务优先级

        Args:
            task_id: 任务ID
            priority: 新的优先级

        Returns:
            是否成功

        Raises:
            TaskError: 更新失败
        """
        try:
            self._request(
                method='PATCH',
                endpoint=f'/tasks/{task_id}/priority',
                data={'priority': priority}
            )
            return True
        except ConnectionError as e:
            raise TaskError(f"更新优先级失败: {str(e)}")

    def cancel_task(self, task_id: str) -> bool:
        """
        取消任务

        Args:
            task_id: 任务ID

        Returns:
            是否成功

        Raises:
            TaskError: 取消失败
        """
        try:
            self._request(
                method='POST',
                endpoint=f'/tasks/{task_id}/cancel'
            )
            return True
        except ConnectionError as e:
            raise TaskError(f"取消任务失败: {str(e)}")

    def upload_result(self, task_id: str, result: Dict[str, Any]) -> bool:
        """
        上传训练结果

        Args:
            task_id: 任务ID
            result: 训练结果

        Returns:
            是否成功

        Raises:
            UploadError: 上传失败
        """
        try:
            self._request(
                method='POST',
                endpoint=f'/tasks/{task_id}/result',
                data={'result': result}
            )
            return True
        except ConnectionError as e:
            raise UploadError(f"上传结果失败: {str(e)}")

    def get_queue_status(self) -> Dict[str, Any]:
        """
        获取队列状态

        Returns:
            队列状态信息
        """
        response = self._request(
            method='GET',
            endpoint='/queue/status'
        )
        return response

    def reorder_queue(self, task_ids: List[str]) -> bool:
        """
        重新排列队列顺序

        Args:
            task_ids: 任务ID列表（按新的顺序）

        Returns:
            是否成功

        Raises:
            TaskError: 重排失败
        """
        try:
            self._request(
                method='POST',
                endpoint='/queue/reorder',
                data={'task_ids': task_ids}
            )
            return True
        except ConnectionError as e:
            raise TaskError(f"重排队列失败: {str(e)}")
