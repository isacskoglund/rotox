// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: telemetry/v1/main.proto

package telemetryv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TransferSubscribeRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TransferSubscribeRequest) Reset() {
	*x = TransferSubscribeRequest{}
	mi := &file_telemetry_v1_main_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransferSubscribeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransferSubscribeRequest) ProtoMessage() {}

func (x *TransferSubscribeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_telemetry_v1_main_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransferSubscribeRequest.ProtoReflect.Descriptor instead.
func (*TransferSubscribeRequest) Descriptor() ([]byte, []int) {
	return file_telemetry_v1_main_proto_rawDescGZIP(), []int{0}
}

type TransferSubscribeResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Events        []*TransferEvent       `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TransferSubscribeResponse) Reset() {
	*x = TransferSubscribeResponse{}
	mi := &file_telemetry_v1_main_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransferSubscribeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransferSubscribeResponse) ProtoMessage() {}

func (x *TransferSubscribeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_telemetry_v1_main_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransferSubscribeResponse.ProtoReflect.Descriptor instead.
func (*TransferSubscribeResponse) Descriptor() ([]byte, []int) {
	return file_telemetry_v1_main_proto_rawDescGZIP(), []int{1}
}

func (x *TransferSubscribeResponse) GetEvents() []*TransferEvent {
	if x != nil {
		return x.Events
	}
	return nil
}

type TransferEvent struct {
	state        protoimpl.MessageState `protogen:"open.v1"`
	ConnectionId string                 `protobuf:"bytes,1,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	// Unix epoch ns
	StartedAt uint64 `protobuf:"varint,2,opt,name=started_at,json=startedAt,proto3" json:"started_at,omitempty"`
	// Unix epoch ns
	FinishedAt    uint64 `protobuf:"varint,3,opt,name=finished_at,json=finishedAt,proto3" json:"finished_at,omitempty"`
	BytesCount    uint64 `protobuf:"varint,4,opt,name=bytes_count,json=bytesCount,proto3" json:"bytes_count,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TransferEvent) Reset() {
	*x = TransferEvent{}
	mi := &file_telemetry_v1_main_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransferEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransferEvent) ProtoMessage() {}

func (x *TransferEvent) ProtoReflect() protoreflect.Message {
	mi := &file_telemetry_v1_main_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransferEvent.ProtoReflect.Descriptor instead.
func (*TransferEvent) Descriptor() ([]byte, []int) {
	return file_telemetry_v1_main_proto_rawDescGZIP(), []int{2}
}

func (x *TransferEvent) GetConnectionId() string {
	if x != nil {
		return x.ConnectionId
	}
	return ""
}

func (x *TransferEvent) GetStartedAt() uint64 {
	if x != nil {
		return x.StartedAt
	}
	return 0
}

func (x *TransferEvent) GetFinishedAt() uint64 {
	if x != nil {
		return x.FinishedAt
	}
	return 0
}

func (x *TransferEvent) GetBytesCount() uint64 {
	if x != nil {
		return x.BytesCount
	}
	return 0
}

type ConnectionSubscribeRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConnectionSubscribeRequest) Reset() {
	*x = ConnectionSubscribeRequest{}
	mi := &file_telemetry_v1_main_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConnectionSubscribeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectionSubscribeRequest) ProtoMessage() {}

func (x *ConnectionSubscribeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_telemetry_v1_main_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectionSubscribeRequest.ProtoReflect.Descriptor instead.
func (*ConnectionSubscribeRequest) Descriptor() ([]byte, []int) {
	return file_telemetry_v1_main_proto_rawDescGZIP(), []int{3}
}

type ConnectionSubscribeResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Events        []*ConnectionEvent     `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConnectionSubscribeResponse) Reset() {
	*x = ConnectionSubscribeResponse{}
	mi := &file_telemetry_v1_main_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConnectionSubscribeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectionSubscribeResponse) ProtoMessage() {}

func (x *ConnectionSubscribeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_telemetry_v1_main_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectionSubscribeResponse.ProtoReflect.Descriptor instead.
func (*ConnectionSubscribeResponse) Descriptor() ([]byte, []int) {
	return file_telemetry_v1_main_proto_rawDescGZIP(), []int{4}
}

func (x *ConnectionSubscribeResponse) GetEvents() []*ConnectionEvent {
	if x != nil {
		return x.Events
	}
	return nil
}

type ConnectionEvent struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ConnectionId  string                 `protobuf:"bytes,1,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	ClientAddress string                 `protobuf:"bytes,2,opt,name=client_address,json=clientAddress,proto3" json:"client_address,omitempty"`
	TargetAddress string                 `protobuf:"bytes,3,opt,name=target_address,json=targetAddress,proto3" json:"target_address,omitempty"`
	// Unix epoch ns
	OpenedAt uint64 `protobuf:"varint,4,opt,name=opened_at,json=openedAt,proto3" json:"opened_at,omitempty"`
	// Unix epoch ns
	// 0 indicates yet to be closed
	ClosedAt      uint64 `protobuf:"varint,5,opt,name=closed_at,json=closedAt,proto3" json:"closed_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConnectionEvent) Reset() {
	*x = ConnectionEvent{}
	mi := &file_telemetry_v1_main_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConnectionEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectionEvent) ProtoMessage() {}

func (x *ConnectionEvent) ProtoReflect() protoreflect.Message {
	mi := &file_telemetry_v1_main_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectionEvent.ProtoReflect.Descriptor instead.
func (*ConnectionEvent) Descriptor() ([]byte, []int) {
	return file_telemetry_v1_main_proto_rawDescGZIP(), []int{5}
}

func (x *ConnectionEvent) GetConnectionId() string {
	if x != nil {
		return x.ConnectionId
	}
	return ""
}

func (x *ConnectionEvent) GetClientAddress() string {
	if x != nil {
		return x.ClientAddress
	}
	return ""
}

func (x *ConnectionEvent) GetTargetAddress() string {
	if x != nil {
		return x.TargetAddress
	}
	return ""
}

func (x *ConnectionEvent) GetOpenedAt() uint64 {
	if x != nil {
		return x.OpenedAt
	}
	return 0
}

func (x *ConnectionEvent) GetClosedAt() uint64 {
	if x != nil {
		return x.ClosedAt
	}
	return 0
}

var File_telemetry_v1_main_proto protoreflect.FileDescriptor

const file_telemetry_v1_main_proto_rawDesc = "" +
	"\n" +
	"\x17telemetry/v1/main.proto\x12\ftelemetry.v1\"\x1a\n" +
	"\x18TransferSubscribeRequest\"P\n" +
	"\x19TransferSubscribeResponse\x123\n" +
	"\x06events\x18\x01 \x03(\v2\x1b.telemetry.v1.TransferEventR\x06events\"\x95\x01\n" +
	"\rTransferEvent\x12#\n" +
	"\rconnection_id\x18\x01 \x01(\tR\fconnectionId\x12\x1d\n" +
	"\n" +
	"started_at\x18\x02 \x01(\x04R\tstartedAt\x12\x1f\n" +
	"\vfinished_at\x18\x03 \x01(\x04R\n" +
	"finishedAt\x12\x1f\n" +
	"\vbytes_count\x18\x04 \x01(\x04R\n" +
	"bytesCount\"\x1c\n" +
	"\x1aConnectionSubscribeRequest\"T\n" +
	"\x1bConnectionSubscribeResponse\x125\n" +
	"\x06events\x18\x01 \x03(\v2\x1d.telemetry.v1.ConnectionEventR\x06events\"\xbe\x01\n" +
	"\x0fConnectionEvent\x12#\n" +
	"\rconnection_id\x18\x01 \x01(\tR\fconnectionId\x12%\n" +
	"\x0eclient_address\x18\x02 \x01(\tR\rclientAddress\x12%\n" +
	"\x0etarget_address\x18\x03 \x01(\tR\rtargetAddress\x12\x1b\n" +
	"\topened_at\x18\x04 \x01(\x04R\bopenedAt\x12\x1b\n" +
	"\tclosed_at\x18\x05 \x01(\x04R\bclosedAt2\xe8\x01\n" +
	"\x10TelemetryService\x12f\n" +
	"\x11TransferSubscribe\x12&.telemetry.v1.TransferSubscribeRequest\x1a'.telemetry.v1.TransferSubscribeResponse0\x01\x12l\n" +
	"\x13ConnectionSubscribe\x12(.telemetry.v1.ConnectionSubscribeRequest\x1a).telemetry.v1.ConnectionSubscribeResponse0\x01B\xa7\x01\n" +
	"\x10com.telemetry.v1B\tMainProtoP\x01Z7github.com/isacskoglund/goroxy/telemetry/v1;telemetryv1\xa2\x02\x03TXX\xaa\x02\fTelemetry.V1\xca\x02\fTelemetry\\V1\xe2\x02\x18Telemetry\\V1\\GPBMetadata\xea\x02\rTelemetry::V1b\x06proto3"

var (
	file_telemetry_v1_main_proto_rawDescOnce sync.Once
	file_telemetry_v1_main_proto_rawDescData []byte
)

func file_telemetry_v1_main_proto_rawDescGZIP() []byte {
	file_telemetry_v1_main_proto_rawDescOnce.Do(func() {
		file_telemetry_v1_main_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_telemetry_v1_main_proto_rawDesc), len(file_telemetry_v1_main_proto_rawDesc)))
	})
	return file_telemetry_v1_main_proto_rawDescData
}

var file_telemetry_v1_main_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_telemetry_v1_main_proto_goTypes = []any{
	(*TransferSubscribeRequest)(nil),    // 0: telemetry.v1.TransferSubscribeRequest
	(*TransferSubscribeResponse)(nil),   // 1: telemetry.v1.TransferSubscribeResponse
	(*TransferEvent)(nil),               // 2: telemetry.v1.TransferEvent
	(*ConnectionSubscribeRequest)(nil),  // 3: telemetry.v1.ConnectionSubscribeRequest
	(*ConnectionSubscribeResponse)(nil), // 4: telemetry.v1.ConnectionSubscribeResponse
	(*ConnectionEvent)(nil),             // 5: telemetry.v1.ConnectionEvent
}
var file_telemetry_v1_main_proto_depIdxs = []int32{
	2, // 0: telemetry.v1.TransferSubscribeResponse.events:type_name -> telemetry.v1.TransferEvent
	5, // 1: telemetry.v1.ConnectionSubscribeResponse.events:type_name -> telemetry.v1.ConnectionEvent
	0, // 2: telemetry.v1.TelemetryService.TransferSubscribe:input_type -> telemetry.v1.TransferSubscribeRequest
	3, // 3: telemetry.v1.TelemetryService.ConnectionSubscribe:input_type -> telemetry.v1.ConnectionSubscribeRequest
	1, // 4: telemetry.v1.TelemetryService.TransferSubscribe:output_type -> telemetry.v1.TransferSubscribeResponse
	4, // 5: telemetry.v1.TelemetryService.ConnectionSubscribe:output_type -> telemetry.v1.ConnectionSubscribeResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_telemetry_v1_main_proto_init() }
func file_telemetry_v1_main_proto_init() {
	if File_telemetry_v1_main_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_telemetry_v1_main_proto_rawDesc), len(file_telemetry_v1_main_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_telemetry_v1_main_proto_goTypes,
		DependencyIndexes: file_telemetry_v1_main_proto_depIdxs,
		MessageInfos:      file_telemetry_v1_main_proto_msgTypes,
	}.Build()
	File_telemetry_v1_main_proto = out.File
	file_telemetry_v1_main_proto_goTypes = nil
	file_telemetry_v1_main_proto_depIdxs = nil
}
