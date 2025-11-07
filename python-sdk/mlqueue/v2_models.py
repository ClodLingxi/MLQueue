"""
MLQueue V2 数据模型
新架构：User -> Group -> TrainingUnit -> TrainingQueue
"""
from typing import Optional, Dict, Any, List
from datetime import datetime
from enum import Enum
import threading
import time


class QueueStatus(Enum):
    """训练队列状态"""
    PENDING = "pending"      # 等待执行
    RUNNING = "running"      # 执行中
    COMPLETED = "completed"  # 已完成
    FAILED = "failed"        # 失败


class CreatedBy(Enum):
    """创建来源"""
    CLIENT = "client"   # Python客户端创建
    WEB = "web"        # Web前端创建


class Group:
    """
    组对象 - 代表一个ML项目
    """

    def __init__(
        self,
        client,
        group_id: str,
        name: str,
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        created_at: Optional[str] = None,
        updated_at: Optional[str] = None
    ):
        """
        初始化组对象

        Args:
            client: MLQueueV2Client实例
            group_id: 组ID
            name: 组名称
            description: 组描述
            metadata: 元数据
            created_at: 创建时间
            updated_at: 更新时间
        """
        self.client = client
        self.id = group_id
        self.name = name
        self.description = description
        self.metadata = metadata or {}
        self.created_at = created_at
        self.updated_at = updated_at

    def create_training_unit(
        self,
        name: str,
        config: Dict[str, Any],
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> 'TrainingUnit':
        """
        在该组下创建训练单元

        Args:
            name: 训练单元名称
            config: 训练配置
            description: 描述
            metadata: 元数据

        Returns:
            TrainingUnit对象
        """
        return self.client.create_training_unit(
            group_id=self.id,
            name=name,
            config=config,
            description=description,
            metadata=metadata
        )

    def list_training_units(self) -> List['TrainingUnit']:
        """
        列出该组下的所有训练单元

        Returns:
            TrainingUnit对象列表
        """
        return self.client.list_training_units(self.id)

    def update(
        self,
        name: Optional[str] = None,
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> 'Group':
        """
        更新组信息

        Args:
            name: 新名称
            description: 新描述
            metadata: 新元数据

        Returns:
            更新后的Group对象
        """
        return self.client.update_group(
            self.id,
            name=name,
            description=description,
            metadata=metadata
        )

    def delete(self) -> bool:
        """
        删除该组

        Returns:
            是否成功
        """
        return self.client.delete_group(self.id)

    def to_dict(self) -> Dict[str, Any]:
        """转换为字典"""
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "metadata": self.metadata,
            "created_at": self.created_at,
            "updated_at": self.updated_at
        }

    def __repr__(self) -> str:
        return f"Group(id={self.id}, name={self.name})"


class TrainingUnit:
    """
    训练单元对象 - 云端和本地各保留一份
    """

    def __init__(
        self,
        client,
        unit_id: str,
        group_id: str,
        name: str,
        config: Dict[str, Any],
        version: int = 1,
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        created_at: Optional[str] = None,
        updated_at: Optional[str] = None
    ):
        """
        初始化训练单元对象

        Args:
            client: MLQueueV2Client实例
            unit_id: 训练单元ID
            group_id: 所属组ID
            name: 训练单元名称
            config: 训练配置
            version: 版本号（用于同步检测）
            description: 描述
            metadata: 元数据
            created_at: 创建时间
            updated_at: 更新时间
        """
        self.client = client
        self.id = unit_id
        self.group_id = group_id
        self.name = name
        self.config = config
        self.version = version
        self.description = description
        self.metadata = metadata or {}
        self.created_at = created_at
        self.updated_at = updated_at
        self._queues: List[TrainingQueue] = []

        # 心跳相关
        self._heartbeat_thread: Optional[threading.Thread] = None
        self._heartbeat_running = False
        self._heartbeat_interval = 6  # 每6秒发送一次心跳（5-8秒范围内）

    def add_queue(
        self,
        name: str,
        parameters: Dict[str, Any],
        created_by: str = "client",
        metadata: Optional[Dict[str, Any]] = None
    ) -> 'TrainingQueue':
        """
        添加训练队列

        Args:
            name: 队列名称
            parameters: 训练参数
            created_by: 创建来源 ("client" 或 "web")
            metadata: 元数据

        Returns:
            TrainingQueue对象
        """
        return self.client.create_queue(
            unit_id=self.id,
            name=name,
            parameters=parameters,
            created_by=created_by,
            metadata=metadata
        )

    def add_queues_batch(
        self,
        queues: List[Dict[str, Any]]
    ) -> List['TrainingQueue']:
        """
        批量添加训练队列

        Args:
            queues: 队列配置列表，每个包含name, parameters等字段

        Returns:
            TrainingQueue对象列表
        """
        return self.client.create_queues_batch(self.id, queues)

    def list_queues(
        self,
        status: Optional[QueueStatus] = None
    ) -> List['TrainingQueue']:
        """
        列出该训练单元的所有队列

        Args:
            status: 可选的状态过滤

        Returns:
            TrainingQueue对象列表
        """
        queues = self.client.list_queues(self.id, status=status)
        self._queues = queues
        return queues

    def get_pending_queues(self) -> List['TrainingQueue']:
        """
        获取所有待执行的队列

        Returns:
            待执行的TrainingQueue对象列表
        """
        return self.list_queues(status=QueueStatus.PENDING)

    def reorder(self, queue_ids: List[str]) -> Dict[str, Any]:
        """
        重新排序训练队列

        注意：只能调整pending状态的队列

        Args:
            queue_ids: 队列ID列表（按期望的执行顺序排列）

        Returns:
            操作结果
        """
        return self.client.reorder_queues(self.id, queue_ids)

    def sync(self) -> Dict[str, Any]:
        """
        主动同步：从云端拉取最新配置

        返回同步结果，包含是否需要同步和最新的队列列表

        Returns:
            同步结果字典
        """
        result = self.client.sync_training_unit(self.id, self.version)

        if result.get("need_sync"):
            # 需要同步，更新本地数据
            self.version = result["server_version"]
            if "queues" in result and result["queues"] is not None:
                self._queues = [
                    TrainingQueue.from_dict(self.client, q)
                    for q in result["queues"]
                ]


        return result

    def update(
        self,
        name: Optional[str] = None,
        config: Optional[Dict[str, Any]] = None,
        description: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> 'TrainingUnit':
        """
        更新训练单元

        Args:
            name: 新名称
            config: 新配置
            description: 新描述
            metadata: 新元数据

        Returns:
            更新后的TrainingUnit对象
        """
        return self.client.update_training_unit(
            self.id,
            name=name,
            config=config,
            description=description,
            metadata=metadata
        )

    def delete(self) -> bool:
        """
        删除该训练单元

        Returns:
            是否成功
        """
        # 停止心跳
        self.stop_heartbeat()
        return self.client.delete_training_unit(self.id)

    def _heartbeat_loop(self):
        """心跳循环（在后台线程中运行）"""
        while self._heartbeat_running:
            try:
                response = self.client.heartbeat(self.id)
                # 可选：记录心跳状态
                # print(f"[心跳] {self.name}: {response.get('connection_status')}")
            except Exception as e:
                # 心跳失败不应中断程序，只记录错误
                print(f"[心跳错误] {self.name}: {str(e)}")

            # 等待下一次心跳
            time.sleep(self._heartbeat_interval)

    def start_heartbeat(self, interval: int = 6):
        """
        启动心跳线程

        Python客户端应在创建训练单元后立即启动心跳，
        以保持与云端的连接状态。心跳会在后台线程中每隔5-8秒自动发送。

        Args:
            interval: 心跳间隔（秒），推荐5-8秒，默认6秒
        """
        if self._heartbeat_running:
            print(f"[心跳] {self.name}: 心跳已在运行")
            return

        self._heartbeat_interval = interval
        self._heartbeat_running = True

        # 创建并启动心跳线程
        self._heartbeat_thread = threading.Thread(
            target=self._heartbeat_loop,
            daemon=True,  # 设置为守护线程，程序退出时自动结束
            name=f"Heartbeat-{self.id}"
        )
        self._heartbeat_thread.start()
        print(f"[心跳] {self.name}: 心跳已启动 (间隔={interval}秒)")

    def stop_heartbeat(self):
        """
        停止心跳线程

        在训练完成或程序退出前应调用此方法停止心跳。
        """
        if not self._heartbeat_running:
            return

        self._heartbeat_running = False

        # 等待心跳线程结束
        if self._heartbeat_thread and self._heartbeat_thread.is_alive():
            self._heartbeat_thread.join(timeout=2)

        print(f"[心跳] {self.name}: 心跳已停止")

    def to_dict(self) -> Dict[str, Any]:
        """转换为字典"""
        return {
            "id": self.id,
            "group_id": self.group_id,
            "name": self.name,
            "config": self.config,
            "version": self.version,
            "description": self.description,
            "metadata": self.metadata,
            "created_at": self.created_at,
            "updated_at": self.updated_at
        }

    def __repr__(self) -> str:
        return f"TrainingUnit(id={self.id}, name={self.name}, version={self.version})"


class TrainingQueue:
    """
    训练队列对象 - 具体的训练任务
    """

    def __init__(
        self,
        client,
        queue_id: str,
        unit_id: str,
        name: str,
        parameters: Dict[str, Any],
        status: str = "pending",
        order: int = 0,
        created_by: str = "client",
        result: Optional[Dict[str, Any]] = None,
        metrics: Optional[Dict[str, Any]] = None,
        error_message: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        started_at: Optional[str] = None,
        completed_at: Optional[str] = None,
        created_at: Optional[str] = None,
        updated_at: Optional[str] = None
    ):
        """
        初始化训练队列对象

        Args:
            client: MLQueueV2Client实例
            queue_id: 队列ID
            unit_id: 所属训练单元ID
            name: 队列名称
            parameters: 训练参数
            status: 状态
            order: 执行顺序（数字越小越先执行）
            created_by: 创建来源
            result: 训练结果
            metrics: 训练指标
            error_message: 错误信息
            metadata: 元数据
            started_at: 开始时间
            completed_at: 完成时间
            created_at: 创建时间
            updated_at: 更新时间
        """
        self.client = client
        self.id = queue_id
        self.unit_id = unit_id
        self.name = name
        self.parameters = parameters
        self.status = QueueStatus(status) if isinstance(status, str) else status
        self.order = order
        self.created_by = created_by
        self.result = result
        self.metrics = metrics
        self.error_message = error_message
        self.metadata = metadata or {}
        self.started_at = started_at
        self.completed_at = completed_at
        self.created_at = created_at
        self.updated_at = updated_at

    def start(self) -> bool:
        """
        开始执行该队列

        Returns:
            是否成功
        """
        success = self.client.start_queue(self.id)
        if success:
            self.status = QueueStatus.RUNNING
            self.started_at = datetime.now().isoformat()
        return success

    def complete(
        self,
        result: Dict[str, Any],
        metrics: Optional[Dict[str, Any]] = None
    ) -> bool:
        """
        标记队列为完成状态

        Args:
            result: 训练结果
            metrics: 训练指标

        Returns:
            是否成功
        """
        success = self.client.complete_queue(self.id, result, metrics)
        if success:
            self.status = QueueStatus.COMPLETED
            self.result = result
            self.metrics = metrics
            self.completed_at = datetime.now().isoformat()
        return success

    def fail(self, error_message: str) -> bool:
        """
        标记队列为失败状态

        Args:
            error_message: 错误信息

        Returns:
            是否成功
        """
        success = self.client.fail_queue(self.id, error_message)
        if success:
            self.status = QueueStatus.FAILED
            self.error_message = error_message
            self.completed_at = datetime.now().isoformat()
        return success

    def update(
        self,
        name: Optional[str] = None,
        parameters: Optional[Dict[str, Any]] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> 'TrainingQueue':
        """
        更新队列信息（仅限pending状态）

        Args:
            name: 新名称
            parameters: 新参数
            metadata: 新元数据

        Returns:
            更新后的TrainingQueue对象
        """
        return self.client.update_queue(
            self.id,
            name=name,
            parameters=parameters,
            metadata=metadata
        )

    def delete(self) -> bool:
        """
        删除该队列

        Returns:
            是否成功
        """
        return self.client.delete_queue(self.id)

    def to_dict(self) -> Dict[str, Any]:
        """转换为字典"""
        return {
            "id": self.id,
            "unit_id": self.unit_id,
            "name": self.name,
            "parameters": self.parameters,
            "status": self.status.value if isinstance(self.status, QueueStatus) else self.status,
            "order": self.order,
            "created_by": self.created_by,
            "result": self.result,
            "metrics": self.metrics,
            "error_message": self.error_message,
            "metadata": self.metadata,
            "started_at": self.started_at,
            "completed_at": self.completed_at,
            "created_at": self.created_at,
            "updated_at": self.updated_at
        }

    @classmethod
    def from_dict(cls, client, data: Dict[str, Any]) -> 'TrainingQueue':
        """从字典创建TrainingQueue对象"""
        # 兼容两种ID字段名
        queue_id = data.get("queue_id") or data.get("id")

        # 兼容两种错误字段名（API使用error_msg）
        error_msg = data.get("error_msg") or data.get("error_message")

        return cls(
            client=client,
            queue_id=queue_id,
            unit_id=data.get("unit_id"),
            name=data.get("name"),
            parameters=data.get("parameters", {}),
            status=data.get("status", "pending"),
            order=data.get("order", 0),
            created_by=data.get("created_by", "client"),
            result=data.get("result"),
            metrics=data.get("metrics"),
            error_message=error_msg,
            metadata=data.get("metadata"),
            started_at=data.get("started_at"),
            completed_at=data.get("completed_at"),
            created_at=data.get("created_at"),
            updated_at=data.get("updated_at")
        )

    def __repr__(self) -> str:
        return f"TrainingQueue(id={self.id}, name={self.name}, status={self.status.value})"
