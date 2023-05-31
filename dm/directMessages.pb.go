////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: directMessages.proto

package dm

import (
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

// Text is the payload for sending normal text messages the replyMessageID
// is nil when it is not a reply.
type Text struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version        uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Text           string `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	ReplyMessageID []byte `protobuf:"bytes,3,opt,name=replyMessageID,proto3" json:"replyMessageID,omitempty"`
}

func (x *Text) Reset() {
	*x = Text{}
	if protoimpl.UnsafeEnabled {
		mi := &file_directMessages_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Text) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Text) ProtoMessage() {}

func (x *Text) ProtoReflect() protoreflect.Message {
	mi := &file_directMessages_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Text.ProtoReflect.Descriptor instead.
func (*Text) Descriptor() ([]byte, []int) {
	return file_directMessages_proto_rawDescGZIP(), []int{0}
}

func (x *Text) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *Text) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *Text) GetReplyMessageID() []byte {
	if x != nil {
		return x.ReplyMessageID
	}
	return nil
}

// Reaction is the payload for reactions. The reaction must be a
// single emoji and the reactionMessageID must be non nil and a real message
// in the channel.
type Reaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version           uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Reaction          string `protobuf:"bytes,2,opt,name=reaction,proto3" json:"reaction,omitempty"`
	ReactionMessageID []byte `protobuf:"bytes,3,opt,name=reactionMessageID,proto3" json:"reactionMessageID,omitempty"`
}

func (x *Reaction) Reset() {
	*x = Reaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_directMessages_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reaction) ProtoMessage() {}

func (x *Reaction) ProtoReflect() protoreflect.Message {
	mi := &file_directMessages_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reaction.ProtoReflect.Descriptor instead.
func (*Reaction) Descriptor() ([]byte, []int) {
	return file_directMessages_proto_rawDescGZIP(), []int{1}
}

func (x *Reaction) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *Reaction) GetReaction() string {
	if x != nil {
		return x.Reaction
	}
	return ""
}

func (x *Reaction) GetReactionMessageID() []byte {
	if x != nil {
		return x.ReactionMessageID
	}
	return nil
}

// ChannelInvitation is the payload for a Invitation MessageType. It the DM
// partner to a channel.
type ChannelInvitation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version    uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Text       string `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	InviteLink string `protobuf:"bytes,3,opt,name=inviteLink,proto3" json:"inviteLink,omitempty"`
	Password   string `protobuf:"bytes,4,opt,name=Password,proto3" json:"Password,omitempty"`
}

func (x *ChannelInvitation) Reset() {
	*x = ChannelInvitation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_directMessages_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChannelInvitation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChannelInvitation) ProtoMessage() {}

func (x *ChannelInvitation) ProtoReflect() protoreflect.Message {
	mi := &file_directMessages_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChannelInvitation.ProtoReflect.Descriptor instead.
func (*ChannelInvitation) Descriptor() ([]byte, []int) {
	return file_directMessages_proto_rawDescGZIP(), []int{2}
}

func (x *ChannelInvitation) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *ChannelInvitation) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *ChannelInvitation) GetInviteLink() string {
	if x != nil {
		return x.InviteLink
	}
	return ""
}

func (x *ChannelInvitation) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

// SilentMessage is the payload for a Silent MessageType. Its primary purpose is
// to communicate new nicknames without sending a Text.
type SilentMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *SilentMessage) Reset() {
	*x = SilentMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_directMessages_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SilentMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SilentMessage) ProtoMessage() {}

func (x *SilentMessage) ProtoReflect() protoreflect.Message {
	mi := &file_directMessages_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SilentMessage.ProtoReflect.Descriptor instead.
func (*SilentMessage) Descriptor() ([]byte, []int) {
	return file_directMessages_proto_rawDescGZIP(), []int{3}
}

func (x *SilentMessage) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

// DirectMessage is a message sent directly from one user to another. It
// includes the return information (public key and dmtoken) for the sender.
type DirectMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The round this message was sent on to the intended recipient
	RoundID uint64 `protobuf:"varint,1,opt,name=RoundID,proto3" json:"RoundID,omitempty"`
	// The round this message was sent on for the self send.
	SelfRoundID uint64 `protobuf:"varint,2,opt,name=SelfRoundID,proto3" json:"SelfRoundID,omitempty"`
	DMToken     uint32 `protobuf:"varint,3,opt,name=DMToken,proto3" json:"DMToken,omitempty"` // hash of private key of the sender
	// The type the below payload is (currently a Text or Reaction)
	PayloadType uint32 `protobuf:"varint,4,opt,name=PayloadType,proto3" json:"PayloadType,omitempty"`
	// Payload is the actual message payload. It will be processed differently
	// based on the PayloadType.
	Payload []byte `protobuf:"bytes,5,opt,name=Payload,proto3" json:"Payload,omitempty"`
	// nickname is the name which the user is using for this message it will not
	// be longer than 24 characters.
	Nickname string `protobuf:"bytes,6,opt,name=Nickname,proto3" json:"Nickname,omitempty"`
	// Nonce is 32 bits of randomness to ensure that two messages in the same
	// round with that have the same nickname, payload, and lease will not have
	// the same message ID.
	Nonce []byte `protobuf:"bytes,7,opt,name=Nonce,proto3" json:"Nonce,omitempty"`
	// LocalTimestamp is the timestamp when the "send call" is made based upon
	// the local clock. If this differs by more than 5 seconds +/- from when the
	// round it sent on is queued, then a random mutation on the queued time
	// (+/- 200ms) will be used by local clients instead.
	LocalTimestamp int64 `protobuf:"varint,8,opt,name=LocalTimestamp,proto3" json:"LocalTimestamp,omitempty"`
}

func (x *DirectMessage) Reset() {
	*x = DirectMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_directMessages_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DirectMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DirectMessage) ProtoMessage() {}

func (x *DirectMessage) ProtoReflect() protoreflect.Message {
	mi := &file_directMessages_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DirectMessage.ProtoReflect.Descriptor instead.
func (*DirectMessage) Descriptor() ([]byte, []int) {
	return file_directMessages_proto_rawDescGZIP(), []int{4}
}

func (x *DirectMessage) GetRoundID() uint64 {
	if x != nil {
		return x.RoundID
	}
	return 0
}

func (x *DirectMessage) GetSelfRoundID() uint64 {
	if x != nil {
		return x.SelfRoundID
	}
	return 0
}

func (x *DirectMessage) GetDMToken() uint32 {
	if x != nil {
		return x.DMToken
	}
	return 0
}

func (x *DirectMessage) GetPayloadType() uint32 {
	if x != nil {
		return x.PayloadType
	}
	return 0
}

func (x *DirectMessage) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *DirectMessage) GetNickname() string {
	if x != nil {
		return x.Nickname
	}
	return ""
}

func (x *DirectMessage) GetNonce() []byte {
	if x != nil {
		return x.Nonce
	}
	return nil
}

func (x *DirectMessage) GetLocalTimestamp() int64 {
	if x != nil {
		return x.LocalTimestamp
	}
	return 0
}

var File_directMessages_proto protoreflect.FileDescriptor

var file_directMessages_proto_rawDesc = []byte{
	0x0a, 0x14, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x64, 0x6d, 0x22, 0x5c, 0x0a, 0x04, 0x54, 0x65,
	0x78, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74,
	0x12, 0x26, 0x0a, 0x0e, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x22, 0x6e, 0x0a, 0x08, 0x52, 0x65, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1a,
	0x0a, 0x08, 0x72, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x72, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x0a, 0x11, 0x72, 0x65,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x11, 0x72, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x22, 0x7d, 0x0a, 0x11, 0x43, 0x68, 0x61, 0x6e,
	0x6e, 0x65, 0x6c, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x69,
	0x6e, 0x76, 0x69, 0x74, 0x65, 0x4c, 0x69, 0x6e, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x4c, 0x69, 0x6e, 0x6b, 0x12, 0x1a, 0x0a, 0x08, 0x50,
	0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x50,
	0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x22, 0x29, 0x0a, 0x0d, 0x53, 0x69, 0x6c, 0x65, 0x6e,
	0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x22, 0xfb, 0x01, 0x0a, 0x0d, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x49, 0x44, 0x12, 0x20,
	0x0a, 0x0b, 0x53, 0x65, 0x6c, 0x66, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x49, 0x44, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x0b, 0x53, 0x65, 0x6c, 0x66, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x49, 0x44,
	0x12, 0x18, 0x0a, 0x07, 0x44, 0x4d, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x07, 0x44, 0x4d, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x50, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x0b, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x50,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x4e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x05, 0x4e, 0x6f, 0x6e, 0x63, 0x65, 0x12, 0x26, 0x0a, 0x0e, 0x4c, 0x6f, 0x63, 0x61,
	0x6c, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x08, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0e, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x42, 0x1e, 0x5a, 0x1c, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65,
	0x6c, 0x69, 0x78, 0x78, 0x69, 0x72, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x64, 0x6d,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_directMessages_proto_rawDescOnce sync.Once
	file_directMessages_proto_rawDescData = file_directMessages_proto_rawDesc
)

func file_directMessages_proto_rawDescGZIP() []byte {
	file_directMessages_proto_rawDescOnce.Do(func() {
		file_directMessages_proto_rawDescData = protoimpl.X.CompressGZIP(file_directMessages_proto_rawDescData)
	})
	return file_directMessages_proto_rawDescData
}

var file_directMessages_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_directMessages_proto_goTypes = []interface{}{
	(*Text)(nil),              // 0: dm.Text
	(*Reaction)(nil),          // 1: dm.Reaction
	(*ChannelInvitation)(nil), // 2: dm.ChannelInvitation
	(*SilentMessage)(nil),     // 3: dm.SilentMessage
	(*DirectMessage)(nil),     // 4: dm.DirectMessage
}
var file_directMessages_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_directMessages_proto_init() }
func file_directMessages_proto_init() {
	if File_directMessages_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_directMessages_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Text); i {
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
		file_directMessages_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reaction); i {
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
		file_directMessages_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChannelInvitation); i {
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
		file_directMessages_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SilentMessage); i {
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
		file_directMessages_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DirectMessage); i {
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
			RawDescriptor: file_directMessages_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_directMessages_proto_goTypes,
		DependencyIndexes: file_directMessages_proto_depIdxs,
		MessageInfos:      file_directMessages_proto_msgTypes,
	}.Build()
	File_directMessages_proto = out.File
	file_directMessages_proto_rawDesc = nil
	file_directMessages_proto_goTypes = nil
	file_directMessages_proto_depIdxs = nil
}
