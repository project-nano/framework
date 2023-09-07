# ChangeLog

## [1.0.10] 2023-09-07

### Changed

- update to go 1.19
- update dependent packages

## [1.0.9] 2022-08-22

### Changed

- Remove debug log in Daemion.isRunning

## [1.0.8] 2022-08-14

### Fixed

- isRunning panic with null process

## [1.0.7] 2021-06-15

### Changed

- Update go.sum

## [1.0.6] 2021-06-15

### Fixed

- Snapshot cause daemon halt

## [1.0.5] 2021-06-14

### Added

- Add IsRunning to SimpleRunner
- Add Snapshot interface to daemon

## [1.0.4] 2021-02-18

### Added

- Request/Response: SearchGuest/ModifyAutoStart

## [1.0.3] 2020-10-31

### Added

- ParamKeyFrom/ParamKeyTo/ParamKeyIndex/ParamKeySecurity/ParamKeyPolicy
- Get/Add/Modify/Remove GuestRule
- Change guest rule order/default action
- Change security policy rule order
- Sync disk/media images

## [1.0.2] 2020-04-12

### Added

- ParamKeyData/ParamKeyDisplay/ParamKeyDevice/ParamKeyProtocol/ParamKeyInterface/ParamKeyAction/ParamKeyTemplate
- Query/Add/Modify/Remove for Cell Storage
- Query/Get/Create/Modify/Delete for System Template
- ResetSecret 

## [1.0.1] 2020-01-02
### Added

- ParamKeyHardware from MAC address 

## 2019-6-23

### Added

- Resource: ResourcePriority/ResourceDiskThreshold/ResourceNetworkThreshold

- Operate: OperateGetBatchStop

- Request/Response: ModifyPriority/ModifyDiskThreshold/ModifyNetworkThreshold/StartBatchStopGuest/GetBatchStopGuest

### Changed

- Generate module name base on a specified interface

## 2019-6-3

### Fixed

- Cache truncated payload for parsing JSON message

## 2019-4-17

### Added

- Add service type: ServiceTypeImage (Image Server)

- Register service handler

## 2019-2-28

### Added

- Message/Runnable/Transaction test

- Modify guest name

- Batch creating/deleting guest

- Message::SetID()

- CloneJsonMessage

### Changed

- Refactor runnable

- Enable concurrent session

- Create pip/pid/log base on abs binary path

- Redirect message to incoming channel when the target is self.

### Fixed

- Endpoint test

- Peer Endpoint keep recover after stopped 

## 2018-11-27

### Added

- Resource: ResourceAddressPool/ResourceAddressRange

- Request/Response: QueryAddressPool/GetAddressPool/CreateAddressPool/DeleteAddressPool/ModifyAddressPool/QueryAddressRange/GetAddressRange/AddAddressRange/RemoveAddressRange

- Key: ParamKeyStart/ParamKeyEnd/ParamKeyMask/ParamKeyAllocate/ParamKeyGateway/ParamKeyServer/ParamKeyAssign/ParamKeyInternal/ParamKeyExternal

- Event: AddressPoolChanged

## 2018-10-22

### Added

- Operate: OperateEnable/OperateDisable/OperateMigrate/OperatePurge

- Request/Response: EnableComputePoolCell/DisableComputePoolCell/MigrateInstance/PurgeInstance/AttachInstance/DetachInstance

- Event: InstanceMigratedEvent/InstancePurgedEvent/InstanceAttachedEvent/InstanceDetachedEvent

- Distinguish connection closed by the user or lost

## 2018-9-28

### Added

- Resource: ResourceStoragePool

- Key: ParamKeyAttach

- Request/Response: QueryStoragePool/GetStoragePool/QueryStoragePoolStatus/GetStoragePoolStatus/CreateStoragePool/DeleteStoragePool/ModifyStoragePool

- Event: ComputeCellDisconnectedEvent/ComputeCellAddedEvent/ComputeCellRemovedEvent

### Changed

- Don't recover stub service when stop endpoint

## 2018-8-21

### Added

- Key: ParamKeyPrevious/ParamKeyNext/ParamKeyCurrent/ParamKeyCreate/ParamKeyModify

- Resource: ResourceSnapshot/ResourceMedia

- Operate: OperateAttach/OperateDetach/OperateCommit/OperatePull/OperateRestore

- Request/Response: InsertMedia/EjectMedia/QuerySnapshot/GetSnapshot/CreateSnapshot/DeleteSnapshot/RestoreSnapshot/CommitSnapshot/PullSnapshot

- Event: MediaAttachedEvent/MediaDetachedEvent/SnapshotResumedEvent

## 2018-8-14

### Added

- Key: 	ParamKeyVersion/ ParamKeyAdmin/ ParamKeyModule

## 2018-7-28

### Added

- Resource: ResourceCore/ResourceMemory/ResourceDisk/ResourceAuth

- Operate: OperateResize/OperateShrink

- Request/Response: ModifyCore/ModifyMemory/ModifyAuth/ResizeDisk/ShrinkDisk

- Key: ParamKeyImmediate/ParamKeySystem

## 2018-7-24

### Added

- AddressChangedEvent/ResourceAddress

## [0.1.2] - 2018-7-16

Author: Akumas

### Modified

- send disconnect to remote when stop gracefully

- detect and reconnect stub services when connection lost

- reduce connection check interval

