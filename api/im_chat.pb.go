// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.3
// source: api/im_chat.proto

//this is the app's name,all proto in this app must use this name as the proto package name

package api

import (
	_ "github.com/chenjie199234/Corelib/pbex"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SendReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Target     string `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	TargetType string `protobuf:"bytes,2,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	Msg        string `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
	Extra      string `protobuf:"bytes,4,opt,name=extra,proto3" json:"extra,omitempty"`
}

func (x *SendReq) Reset() {
	*x = SendReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendReq) ProtoMessage() {}

func (x *SendReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendReq.ProtoReflect.Descriptor instead.
func (*SendReq) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{0}
}

func (x *SendReq) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *SendReq) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *SendReq) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *SendReq) GetExtra() string {
	if x != nil {
		return x.Extra
	}
	return ""
}

type SendResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// msg index start from 1
	MsgIndex uint64 `protobuf:"varint,1,opt,name=msg_index,json=msgIndex,proto3" json:"msg_index,omitempty"`
	// unit: nanosecond
	Timestamp uint64 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *SendResp) Reset() {
	*x = SendResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendResp) ProtoMessage() {}

func (x *SendResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendResp.ProtoReflect.Descriptor instead.
func (*SendResp) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{1}
}

func (x *SendResp) GetMsgIndex() uint64 {
	if x != nil {
		return x.MsgIndex
	}
	return 0
}

func (x *SendResp) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type RecallReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Target     string `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	TargetType string `protobuf:"bytes,2,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	MsgIndex   uint64 `protobuf:"varint,3,opt,name=msg_index,json=msgIndex,proto3" json:"msg_index,omitempty"`
}

func (x *RecallReq) Reset() {
	*x = RecallReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecallReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecallReq) ProtoMessage() {}

func (x *RecallReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecallReq.ProtoReflect.Descriptor instead.
func (*RecallReq) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{2}
}

func (x *RecallReq) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *RecallReq) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *RecallReq) GetMsgIndex() uint64 {
	if x != nil {
		return x.MsgIndex
	}
	return 0
}

type RecallResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RecallIndex uint64 `protobuf:"varint,1,opt,name=recall_index,json=recallIndex,proto3" json:"recall_index,omitempty"`
}

func (x *RecallResp) Reset() {
	*x = RecallResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecallResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecallResp) ProtoMessage() {}

func (x *RecallResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecallResp.ProtoReflect.Descriptor instead.
func (*RecallResp) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{3}
}

func (x *RecallResp) GetRecallIndex() uint64 {
	if x != nil {
		return x.RecallIndex
	}
	return 0
}

type AckReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Target     string `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	TargetType string `protobuf:"bytes,2,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	MsgIndex   uint64 `protobuf:"varint,3,opt,name=msg_index,json=msgIndex,proto3" json:"msg_index,omitempty"`
}

func (x *AckReq) Reset() {
	*x = AckReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AckReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AckReq) ProtoMessage() {}

func (x *AckReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AckReq.ProtoReflect.Descriptor instead.
func (*AckReq) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{4}
}

func (x *AckReq) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *AckReq) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *AckReq) GetMsgIndex() uint64 {
	if x != nil {
		return x.MsgIndex
	}
	return 0
}

type AckResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AckResp) Reset() {
	*x = AckResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AckResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AckResp) ProtoMessage() {}

func (x *AckResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AckResp.ProtoReflect.Descriptor instead.
func (*AckResp) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{5}
}

type MsgInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// msg index start from 1
	MsgIndex uint64 `protobuf:"varint,1,opt,name=msg_index,json=msgIndex,proto3" json:"msg_index,omitempty"`
	// recall index start 1,0 means didn't recall
	RecallIndex uint64 `protobuf:"varint,2,opt,name=recall_index,json=recallIndex,proto3" json:"recall_index,omitempty"`
	Msg         string `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
	Extra       string `protobuf:"bytes,4,opt,name=extra,proto3" json:"extra,omitempty"`
	Timestamp   uint64 `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Sender      string `protobuf:"bytes,6,opt,name=sender,proto3" json:"sender,omitempty"`
	Target      string `protobuf:"bytes,7,opt,name=target,proto3" json:"target,omitempty"`
	TargetType  string `protobuf:"bytes,8,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
}

func (x *MsgInfo) Reset() {
	*x = MsgInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgInfo) ProtoMessage() {}

func (x *MsgInfo) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgInfo.ProtoReflect.Descriptor instead.
func (*MsgInfo) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{6}
}

func (x *MsgInfo) GetMsgIndex() uint64 {
	if x != nil {
		return x.MsgIndex
	}
	return 0
}

func (x *MsgInfo) GetRecallIndex() uint64 {
	if x != nil {
		return x.RecallIndex
	}
	return 0
}

func (x *MsgInfo) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *MsgInfo) GetExtra() string {
	if x != nil {
		return x.Extra
	}
	return ""
}

func (x *MsgInfo) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *MsgInfo) GetSender() string {
	if x != nil {
		return x.Sender
	}
	return ""
}

func (x *MsgInfo) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *MsgInfo) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

type PullReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Target     string `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	TargetType string `protobuf:"bytes,2,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	Direction  string `protobuf:"bytes,3,opt,name=direction,proto3" json:"direction,omitempty"`
	// the response will include this index
	StartMsgIndex    uint64 `protobuf:"varint,4,opt,name=start_msg_index,json=startMsgIndex,proto3" json:"start_msg_index,omitempty"`          //if this is 0,means don't need to pull the msgs
	StartRecallIndex uint64 `protobuf:"varint,5,opt,name=start_recall_index,json=startRecallIndex,proto3" json:"start_recall_index,omitempty"` //if this is 0,means don't need to pull the recalls
	Count            uint64 `protobuf:"varint,6,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *PullReq) Reset() {
	*x = PullReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PullReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullReq) ProtoMessage() {}

func (x *PullReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullReq.ProtoReflect.Descriptor instead.
func (*PullReq) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{7}
}

func (x *PullReq) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *PullReq) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *PullReq) GetDirection() string {
	if x != nil {
		return x.Direction
	}
	return ""
}

func (x *PullReq) GetStartMsgIndex() uint64 {
	if x != nil {
		return x.StartMsgIndex
	}
	return 0
}

func (x *PullReq) GetStartRecallIndex() uint64 {
	if x != nil {
		return x.StartRecallIndex
	}
	return 0
}

func (x *PullReq) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

type PullResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msgs    []*MsgInfo        `protobuf:"bytes,1,rep,name=msgs,proto3" json:"msgs,omitempty"`
	Recalls map[uint64]uint64 `protobuf:"bytes,2,rep,name=recalls,proto3" json:"recalls,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (x *PullResp) Reset() {
	*x = PullResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_im_chat_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PullResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullResp) ProtoMessage() {}

func (x *PullResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_im_chat_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullResp.ProtoReflect.Descriptor instead.
func (*PullResp) Descriptor() ([]byte, []int) {
	return file_api_im_chat_proto_rawDescGZIP(), []int{8}
}

func (x *PullResp) GetMsgs() []*MsgInfo {
	if x != nil {
		return x.Msgs
	}
	return nil
}

func (x *PullResp) GetRecalls() map[uint64]uint64 {
	if x != nil {
		return x.Recalls
	}
	return nil
}

var File_api_im_chat_proto protoreflect.FileDescriptor

var file_api_im_chat_proto_rawDesc = []byte{
	0x0a, 0x11, 0x61, 0x70, 0x69, 0x2f, 0x69, 0x6d, 0x5f, 0x63, 0x68, 0x61, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x02, 0x69, 0x6d, 0x1a, 0x0f, 0x70, 0x62, 0x65, 0x78, 0x2f, 0x70, 0x62,
	0x65, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8a, 0x01, 0x0a, 0x08, 0x73, 0x65, 0x6e,
	0x64, 0x5f, 0x72, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0xf0, 0x90, 0x4e, 0x00, 0x52, 0x06, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x12, 0x32, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x11, 0xd2, 0x90, 0x4e, 0x04, 0x75, 0x73,
	0x65, 0x72, 0xd2, 0x90, 0x4e, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x0a, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0xf0, 0x90, 0x4e, 0x00, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x65, 0x78, 0x74, 0x72, 0x61, 0x22, 0x46, 0x0a, 0x09, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x6d, 0x73, 0x67, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x7b, 0x0a,
	0x0a, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x06, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0xf0, 0x90, 0x4e,
	0x00, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x12, 0x32, 0x0a, 0x0b, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x11,
	0xd2, 0x90, 0x4e, 0x04, 0x75, 0x73, 0x65, 0x72, 0xd2, 0x90, 0x4e, 0x05, 0x67, 0x72, 0x6f, 0x75,
	0x70, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1b, 0x0a,
	0x09, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x08, 0x6d, 0x73, 0x67, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x22, 0x30, 0x0a, 0x0b, 0x72, 0x65,
	0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x65, 0x63,
	0x61, 0x6c, 0x6c, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0b, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x22, 0x78, 0x0a, 0x07,
	0x61, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0xf0, 0x90, 0x4e, 0x00, 0x52, 0x06, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x12, 0x32, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x11, 0xd2, 0x90, 0x4e, 0x04,
	0x75, 0x73, 0x65, 0x72, 0xd2, 0x90, 0x4e, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x0a, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x73, 0x67,
	0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x6d, 0x73,
	0x67, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x22, 0x0a, 0x0a, 0x08, 0x61, 0x63, 0x6b, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x22, 0xe1, 0x01, 0x0a, 0x08, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x12,
	0x1b, 0x0a, 0x09, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x08, 0x6d, 0x73, 0x67, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x21, 0x0a, 0x0c,
	0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0b, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12,
	0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73,
	0x67, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x16, 0x0a,
	0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x22, 0x85, 0x02, 0x0a, 0x08, 0x70, 0x75, 0x6c, 0x6c, 0x5f,
	0x72, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x04, 0xf0, 0x90, 0x4e, 0x00, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x12, 0x32, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x11, 0xd2, 0x90, 0x4e, 0x04, 0x75, 0x73, 0x65, 0x72,
	0xd2, 0x90, 0x4e, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x31, 0x0a, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x13, 0xd2, 0x90, 0x4e, 0x06, 0x62, 0x65,
	0x66, 0x6f, 0x72, 0x65, 0xd2, 0x90, 0x4e, 0x05, 0x61, 0x66, 0x74, 0x65, 0x72, 0x52, 0x09, 0x64,
	0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x74, 0x61, 0x72,
	0x74, 0x5f, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x0d, 0x73, 0x74, 0x61, 0x72, 0x74, 0x4d, 0x73, 0x67, 0x49, 0x6e, 0x64, 0x65, 0x78,
	0x12, 0x2c, 0x0a, 0x12, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c,
	0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x10, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x1e,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x42, 0x08, 0xa0,
	0x91, 0x4e, 0x00, 0xb8, 0x91, 0x4e, 0x32, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x9f,
	0x01, 0x0a, 0x09, 0x70, 0x75, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x12, 0x20, 0x0a, 0x04,
	0x6d, 0x73, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x69, 0x6d, 0x2e,
	0x6d, 0x73, 0x67, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x12, 0x34,
	0x0a, 0x07, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x69, 0x6d, 0x2e, 0x70, 0x75, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x2e, 0x52,
	0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x72, 0x65, 0x63,
	0x61, 0x6c, 0x6c, 0x73, 0x1a, 0x3a, 0x0a, 0x0c, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x32, 0xe9, 0x01, 0x0a, 0x04, 0x63, 0x68, 0x61, 0x74, 0x12, 0x36, 0x0a, 0x04, 0x73, 0x65, 0x6e,
	0x64, 0x12, 0x0c, 0x2e, 0x69, 0x6d, 0x2e, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x1a,
	0x0d, 0x2e, 0x69, 0x6d, 0x2e, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x22, 0x11,
	0x8a, 0x9f, 0x49, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x92, 0x9f, 0x49, 0x05, 0x74, 0x6f, 0x6b, 0x65,
	0x6e, 0x12, 0x3c, 0x0a, 0x06, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x12, 0x0e, 0x2e, 0x69, 0x6d,
	0x2e, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x71, 0x1a, 0x0f, 0x2e, 0x69, 0x6d,
	0x2e, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x22, 0x11, 0x8a, 0x9f,
	0x49, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x92, 0x9f, 0x49, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12,
	0x33, 0x0a, 0x03, 0x61, 0x63, 0x6b, 0x12, 0x0b, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x63, 0x6b, 0x5f,
	0x72, 0x65, 0x71, 0x1a, 0x0c, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x73,
	0x70, 0x22, 0x11, 0x8a, 0x9f, 0x49, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x92, 0x9f, 0x49, 0x05, 0x74,
	0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x36, 0x0a, 0x04, 0x70, 0x75, 0x6c, 0x6c, 0x12, 0x0c, 0x2e, 0x69,
	0x6d, 0x2e, 0x70, 0x75, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x71, 0x1a, 0x0d, 0x2e, 0x69, 0x6d, 0x2e,
	0x70, 0x75, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x22, 0x11, 0x8a, 0x9f, 0x49, 0x04, 0x70,
	0x6f, 0x73, 0x74, 0x92, 0x9f, 0x49, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x42, 0x25, 0x5a, 0x23,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x68, 0x65, 0x6e, 0x6a,
	0x69, 0x65, 0x31, 0x39, 0x39, 0x32, 0x33, 0x34, 0x2f, 0x69, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x3b,
	0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_im_chat_proto_rawDescOnce sync.Once
	file_api_im_chat_proto_rawDescData = file_api_im_chat_proto_rawDesc
)

func file_api_im_chat_proto_rawDescGZIP() []byte {
	file_api_im_chat_proto_rawDescOnce.Do(func() {
		file_api_im_chat_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_im_chat_proto_rawDescData)
	})
	return file_api_im_chat_proto_rawDescData
}

var file_api_im_chat_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_api_im_chat_proto_goTypes = []interface{}{
	(*SendReq)(nil),    // 0: im.send_req
	(*SendResp)(nil),   // 1: im.send_resp
	(*RecallReq)(nil),  // 2: im.recall_req
	(*RecallResp)(nil), // 3: im.recall_resp
	(*AckReq)(nil),     // 4: im.ack_req
	(*AckResp)(nil),    // 5: im.ack_resp
	(*MsgInfo)(nil),    // 6: im.msg_info
	(*PullReq)(nil),    // 7: im.pull_req
	(*PullResp)(nil),   // 8: im.pull_resp
	nil,                // 9: im.pull_resp.RecallsEntry
}
var file_api_im_chat_proto_depIdxs = []int32{
	6, // 0: im.pull_resp.msgs:type_name -> im.msg_info
	9, // 1: im.pull_resp.recalls:type_name -> im.pull_resp.RecallsEntry
	0, // 2: im.chat.send:input_type -> im.send_req
	2, // 3: im.chat.recall:input_type -> im.recall_req
	4, // 4: im.chat.ack:input_type -> im.ack_req
	7, // 5: im.chat.pull:input_type -> im.pull_req
	1, // 6: im.chat.send:output_type -> im.send_resp
	3, // 7: im.chat.recall:output_type -> im.recall_resp
	5, // 8: im.chat.ack:output_type -> im.ack_resp
	8, // 9: im.chat.pull:output_type -> im.pull_resp
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_im_chat_proto_init() }
func file_api_im_chat_proto_init() {
	if File_api_im_chat_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_im_chat_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecallReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecallResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AckReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AckResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PullReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_im_chat_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PullResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_im_chat_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_im_chat_proto_goTypes,
		DependencyIndexes: file_api_im_chat_proto_depIdxs,
		MessageInfos:      file_api_im_chat_proto_msgTypes,
	}.Build()
	File_api_im_chat_proto = out.File
	file_api_im_chat_proto_rawDesc = nil
	file_api_im_chat_proto_goTypes = nil
	file_api_im_chat_proto_depIdxs = nil
}
