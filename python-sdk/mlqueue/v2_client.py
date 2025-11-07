"""
MLQueue V2 API 客户端
Python驱动架构：客户端控制训练执行，云端管理配置
"""
from typing import Optional, Dict, Any, List
import requests
import json

from .v2_models import Group, TrainingUnit, TrainingQueue, QueueStatus
from .exceptions import (
    ConnectionError,
    AuthenticationError,
    TaskError
)


class MLQueueV2Client:
    """
    MLQueue V2 API客户端

    V2架构特点：
    - Python客户端驱动（控制队列执行）
    - 云端管理配置（创建/修改队列参数）
    - 主动同步机制（客户端拉取云端配置）
    """

    def __init__(
        self,
        api_url: str,
        api_key: str,
        timeout: int = 30
    ):
        """
        初始化V2客户端

        Args:
            api_url: V2 API基础URL (例如: http://localhost:8080/v2)
            api_key: API密钥
            timeout: 请求超时时间（秒）
        """
        self.api_url = api_url.rstrip('/')
        self.api_key = api_key
        self.timeout = timeout
        self.session = requests.Session()
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
                error_msg = response.json().get('error', response.text) if response.text else f"HTTP {response.status_code}"
                raise ConnectionError(f"请求失败: {error_msg}")

            return response.json()

        except requests.exceptions.Timeout:
            raise ConnectionError(f"请求超时（{self.timeout}秒）")
        except requests.exceptions.ConnectionError as e:
            raise ConnectionError(f"无法连接到云端服务: {str(e)}")
        except json.JSONDecodeError:
            raise ConnectionError("无效的响应格式")

    # ==================== 组管理 ====================

    def create_group(
        self,
        name: str,
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> Group:
        """
        创建组（代表一个ML项目）

        Args:
            name: 组名称
            description: 组描述
            metadata: 元数据

        Returns:
            Group对象
        """
        data = {
            "name": name,
            "description": description,
            "metadata": metadata or {}
        }

        response = self._request('POST', '/groups', data=data)
        return Group(
            client=self,
            group_id=response['group_id'],
            name=response['name'],
            description=response.get('description'),
            metadata=response.get('metadata'),
            created_at=response.get('created_at'),
            updated_at=response.get('updated_at')
        )

    def list_groups(self) -> List[Group]:
        """
        列出所有组

        Returns:
            Group对象列表
        """
        response = self._request('GET', '/groups')
        groups = response.get('groups', [])
        return [
            Group(
                client=self,
                group_id=g['group_id'],
                name=g['name'],
                description=g.get('description'),
                metadata=g.get('metadata'),
                created_at=g.get('created_at'),
                updated_at=g.get('updated_at')
            )
            for g in groups
        ]

    def get_group(self, group_id: str) -> Group:
        """
        获取组详情

        Args:
            group_id: 组ID

        Returns:
            Group对象
        """
        response = self._request('GET', f'/groups/{group_id}')
        group_data = response['group']
        return Group(
            client=self,
            group_id=group_data['group_id'],
            name=group_data['name'],
            description=group_data.get('description'),
            metadata=group_data.get('metadata'),
            created_at=group_data.get('created_at'),
            updated_at=group_data.get('updated_at')
        )

    def update_group(
        self,
        group_id: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> Group:
        """
        更新组信息

        Args:
            group_id: 组ID
            name: 新名称
            description: 新描述
            metadata: 新元数据

        Returns:
            更新后的Group对象
        """
        data = {}
        if name is not None:
            data['name'] = name
        if description is not None:
            data['description'] = description
        if metadata is not None:
            data['metadata'] = metadata

        response = self._request('PUT', f'/groups/{group_id}', data=data)
        return Group(
            client=self,
            group_id=response['group_id'],
            name=response['name'],
            description=response.get('description'),
            metadata=response.get('metadata'),
            created_at=response.get('created_at'),
            updated_at=response.get('updated_at')
        )

    def delete_group(self, group_id: str) -> bool:
        """
        删除组

        Args:
            group_id: 组ID

        Returns:
            是否成功
        """
        self._request('DELETE', f'/groups/{group_id}')
        return True

    # ==================== 训练单元管理 ====================

    def create_training_unit(
        self,
        group_id: str,
        name: str,
        config: Dict[str, Any],
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> TrainingUnit:
        """
        创建训练单元

        Args:
            group_id: 所属组ID
            name: 训练单元名称
            config: 训练配置
            description: 描述
            metadata: 元数据

        Returns:
            TrainingUnit对象
        """
        data = {
            "name": name,
            "config": config,
            "description": description,
            "metadata": metadata or {}
        }

        response = self._request('POST', f'/groups/{group_id}/units', data=data)

        # 注意：创建时只返回 unit_id 和 version，需要再次获取完整信息
        unit_id = response['unit_id']
        version = response.get('version', 1)

        return TrainingUnit(
            client=self,
            unit_id=unit_id,
            group_id=group_id,
            name=name,
            config=config,
            version=version,
            description=description,
            metadata=metadata or {},
            created_at=None,
            updated_at=None
        )

    def list_training_units(self, group_id: str) -> List[TrainingUnit]:
        """
        列出组下的所有训练单元

        Args:
            group_id: 组ID

        Returns:
            TrainingUnit对象列表
        """
        response = self._request('GET', f'/groups/{group_id}/units')
        units = response.get('units', [])
        return [
            TrainingUnit(
                client=self,
                unit_id=u['unit_id'],
                group_id=u['group_id'],
                name=u['name'],
                config=u['config'],
                version=u.get('version', 1),
                description=u.get('description'),
                metadata=u.get('metadata'),
                created_at=u.get('created_at'),
                updated_at=u.get('updated_at')
            )
            for u in units
        ]

    def get_training_unit(self, unit_id: str) -> TrainingUnit:
        """
        获取训练单元详情

        Args:
            unit_id: 训练单元ID

        Returns:
            TrainingUnit对象
        """
        response = self._request('GET', f'/units/{unit_id}')
        unit_data = response['unit']
        return TrainingUnit(
            client=self,
            unit_id=unit_data['unit_id'],
            group_id=unit_data['group_id'],
            name=unit_data['name'],
            config=unit_data['config'],
            version=unit_data.get('version', 1),
            description=unit_data.get('description'),
            metadata=unit_data.get('metadata'),
            created_at=unit_data.get('created_at'),
            updated_at=unit_data.get('updated_at')
        )

    def update_training_unit(
        self,
        unit_id: str,
        name: Optional[str] = None,
        config: Optional[Dict[str, Any]] = None,
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> TrainingUnit:
        """
        更新训练单元

        Args:
            unit_id: 训练单元ID
            name: 新名称
            config: 新配置
            description: 新描述
            metadata: 新元数据

        Returns:
            更新后的TrainingUnit对象
        """
        data = {}
        if name is not None:
            data['name'] = name
        if config is not None:
            data['config'] = config
        if description is not None:
            data['description'] = description
        if metadata is not None:
            data['metadata'] = metadata

        response = self._request('PUT', f'/units/{unit_id}', data=data)
        return TrainingUnit(
            client=self,
            unit_id=response['unit_id'],
            group_id=response['group_id'],
            name=response['name'],
            config=response['config'],
            version=response.get('version', 1),
            description=response.get('description'),
            metadata=response.get('metadata'),
            created_at=response.get('created_at'),
            updated_at=response.get('updated_at')
        )

    def delete_training_unit(self, unit_id: str) -> bool:
        """
        删除训练单元

        Args:
            unit_id: 训练单元ID

        Returns:
            是否成功
        """
        self._request('DELETE', f'/units/{unit_id}')
        return True

    def sync_training_unit(
        self,
        unit_id: str,
        client_version: int
    ) -> Dict[str, Any]:
        """
        主动同步：从云端拉取最新配置

        Args:
            unit_id: 训练单元ID
            client_version: 客户端当前版本号

        Returns:
            同步结果，包含：
            - need_sync: 是否需要同步
            - cloud_version: 云端版本号
            - unit: 训练单元最新数据
            - queues: 队列列表（如果需要同步）
        """
        data = {"client_version": client_version}
        response = self._request('POST', f'/units/{unit_id}/sync', data=data)

        # 将 cloud_version 映射为 server_version 以保持兼容性
        if 'cloud_version' in response:
            response['server_version'] = response['cloud_version']

        return response

    def heartbeat(self, unit_id: str) -> Dict[str, Any]:
        """
        发送心跳保持连接状态（Python客户端调用）

        Python客户端应每隔5-8秒调用一次此接口。
        如果超过10秒未收到心跳，系统会将连接状态设为disconnected。

        Args:
            unit_id: 训练单元ID

        Returns:
            心跳响应，包含：
            - success: 是否成功
            - connection_status: 连接状态 ("connected" 或 "disconnected")
            - last_heartbeat: 最后心跳时间
        """
        response = self._request('POST', f'/units/{unit_id}/heartbeat')
        return response

    # ==================== 训练队列管理 ====================

    def create_queue(
        self,
        unit_id: str,
        name: str,
        parameters: Dict[str, Any],
        created_by: str = "client",
        metadata: Optional[Dict[str, Any]] = None
    ) -> TrainingQueue:
        """
        创建训练队列

        注意：order字段由服务器自动分配，新队列会追加到末尾

        Args:
            unit_id: 所属训练单元ID
            name: 队列名称
            parameters: 训练参数
            created_by: 创建来源 ("client" 或 "web")
            metadata: 元数据

        Returns:
            TrainingQueue对象
        """
        data = {
            "name": name,
            "parameters": parameters,
            "created_by": created_by,
            "metadata": metadata or {}
        }

        response = self._request('POST', f'/units/{unit_id}/queues', data=data)
        queue_data = response.get('queue', response)  # 兼容两种响应格式
        return TrainingQueue.from_dict(self, queue_data)

    def create_queues_batch(
        self,
        unit_id: str,
        queues: List[Dict[str, Any]],
        created_by: str = "client"
    ) -> List[TrainingQueue]:
        """
        批量创建训练队列

        Args:
            unit_id: 所属训练单元ID
            queues: 队列配置列表
            created_by: 创建来源

        Returns:
            TrainingQueue对象列表
        """
        data = {
            "queues": queues,
            "created_by": created_by
        }
        response = self._request('POST', f'/units/{unit_id}/queues/batch', data=data)

        # 如果只返回了 queue_ids，需要逐个获取详情
        if 'queue_ids' in response:
            queue_ids = response['queue_ids']
            return [self.get_queue(qid) for qid in queue_ids]

        # 如果返回了完整队列数据
        queue_list = response.get('queues', [])
        return [TrainingQueue.from_dict(self, q) for q in queue_list]

    def list_queues(
        self,
        unit_id: str,
        status: Optional[QueueStatus] = None
    ) -> List[TrainingQueue]:
        """
        列出训练单元的所有队列

        Args:
            unit_id: 训练单元ID
            status: 可选的状态过滤

        Returns:
            TrainingQueue对象列表
        """
        params = {}
        if status:
            params['status'] = status.value

        response = self._request('GET', f'/units/{unit_id}/queues', params=params)
        queues = response.get('queues', [])
        return [TrainingQueue.from_dict(self, q) for q in queues]

    def get_queue(self, queue_id: str) -> TrainingQueue:
        """
        获取队列详情

        Args:
            queue_id: 队列ID

        Returns:
            TrainingQueue对象
        """
        response = self._request('GET', f'/queues/{queue_id}')
        queue_data = response['queue']
        return TrainingQueue.from_dict(self, queue_data)

    def update_queue(
        self,
        queue_id: str,
        name: Optional[str] = None,
        parameters: Optional[Dict[str, Any]] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> TrainingQueue:
        """
        更新队列（仅限pending状态）

        注意：order字段不能通过此接口修改，请使用 reorder_queues() 方法

        Args:
            queue_id: 队列ID
            name: 新名称
            parameters: 新参数
            metadata: 新元数据

        Returns:
            更新后的TrainingQueue对象
        """
        data = {}
        if name is not None:
            data['name'] = name
        if parameters is not None:
            data['parameters'] = parameters
        if metadata is not None:
            data['metadata'] = metadata

        response = self._request('PUT', f'/queues/{queue_id}', data=data)
        queue_data = response.get('queue', response)  # 兼容两种响应格式
        return TrainingQueue.from_dict(self, queue_data)

    def delete_queue(self, queue_id: str) -> bool:
        """
        删除队列

        Args:
            queue_id: 队列ID

        Returns:
            是否成功
        """
        self._request('DELETE', f'/queues/{queue_id}')
        return True

    def reorder_queues(
        self,
        unit_id: str,
        queue_ids: List[str]
    ) -> Dict[str, Any]:
        """
        重新排序训练队列

        注意：
        - 只能调整pending状态的队列
        - 队列会按照queue_ids数组顺序重新分配order值

        Args:
            unit_id: 训练单元ID
            queue_ids: 队列ID列表（按期望的执行顺序排列）

        Returns:
            操作结果
        """
        data = {"queue_ids": queue_ids}
        response = self._request('POST', f'/units/{unit_id}/queues/reorder', data=data)
        return response

    # ==================== Python客户端专用 ====================

    def start_queue(self, queue_id: str) -> bool:
        """
        开始执行队列（Python客户端调用）

        Args:
            queue_id: 队列ID

        Returns:
            是否成功
        """
        self._request('POST', f'/queues/{queue_id}/start')
        return True

    def complete_queue(
        self,
        queue_id: str,
        result: Dict[str, Any],
        metrics: Optional[Dict[str, Any]] = None
    ) -> bool:
        """
        标记队列为完成状态（Python客户端调用）

        Args:
            queue_id: 队列ID
            result: 训练结果
            metrics: 训练指标

        Returns:
            是否成功
        """
        data = {
            "result": result,
            "metrics": metrics or {}
        }
        self._request('POST', f'/queues/{queue_id}/complete', data=data)
        return True

    def fail_queue(self, queue_id: str, error_msg: str) -> bool:
        """
        标记队列为失败状态（Python客户端调用）

        Args:
            queue_id: 队列ID
            error_msg: 错误信息

        Returns:
            是否成功
        """
        data = {"error_msg": error_msg}
        self._request('POST', f'/queues/{queue_id}/fail', data=data)
        return True
