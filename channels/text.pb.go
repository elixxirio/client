///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        (unknown)
// source: text.proto

package channels

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

type CMIXChannelText struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version        uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Text           string `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	ReplyMessageID []byte `protobuf:"bytes,3,opt,name=replyMessageID,proto3" json:"replyMessageID,omitempty"`
}

func (x *CMIXChannelText) Reset() {
	*x = CMIXChannelText{}
	if protoimpl.UnsafeEnabled {
		mi := &file_text_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CMIXChannelText) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CMIXChannelText) ProtoMessage() {}

func (x *CMIXChannelText) ProtoReflect() protoreflect.Message {
	mi := &file_text_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CMIXChannelText.ProtoReflect.Descriptor instead.
func (*CMIXChannelText) Descriptor() ([]byte, []int) {
	return file_text_proto_rawDescGZIP(), []int{0}
}

func (x *CMIXChannelText) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *CMIXChannelText) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *CMIXChannelText) GetReplyMessageID() []byte {
	if x != nil {
		return x.ReplyMessageID
	}
	return nil
}

type CMIXChannelReaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version           uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Reaction          uint32 `protobuf:"varint,2,opt,name=reaction,proto3" json:"reaction,omitempty"`
	ReactionMessageID []byte `protobuf:"bytes,3,opt,name=reactionMessageID,proto3" json:"reactionMessageID,omitempty"`
}

func (x *CMIXChannelReaction) Reset() {
	*x = CMIXChannelReaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_text_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CMIXChannelReaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CMIXChannelReaction) ProtoMessage() {}

func (x *CMIXChannelReaction) ProtoReflect() protoreflect.Message {
	mi := &file_text_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CMIXChannelReaction.ProtoReflect.Descriptor instead.
func (*CMIXChannelReaction) Descriptor() ([]byte, []int) {
	return file_text_proto_rawDescGZIP(), []int{1}
}

func (x *CMIXChannelReaction) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *CMIXChannelReaction) GetReaction() uint32 {
	if x != nil {
		return x.Reaction
	}
	return 0
}

func (x *CMIXChannelReaction) GetReactionMessageID() []byte {
	if x != nil {
		return x.ReactionMessageID
	}
	return nil
}

var File_text_proto protoreflect.FileDescriptor

var file_text_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x61,
	0x72, 0x73, 0x65, 0x22, 0x67, 0x0a, 0x0f, 0x43, 0x4d, 0x49, 0x58, 0x43, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c, 0x54, 0x65, 0x78, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x65, 0x78, 0x74, 0x12, 0x26, 0x0a, 0x0e, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x72, 0x65,
	0x70, 0x6c, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x22, 0x79, 0x0a, 0x13,
	0x43, 0x4d, 0x49, 0x58, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a,
	0x08, 0x72, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x08, 0x72, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x0a, 0x11, 0x72, 0x65, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x11, 0x72, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x42, 0x0b, 0x5a, 0x09, 0x2f, 0x63, 0x68, 0x61, 0x6e,
	0x6e, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_text_proto_rawDescOnce sync.Once
	file_text_proto_rawDescData = file_text_proto_rawDesc
)

func file_text_proto_rawDescGZIP() []byte {
	file_text_proto_rawDescOnce.Do(func() {
		file_text_proto_rawDescData = protoimpl.X.CompressGZIP(file_text_proto_rawDescData)
	})
	return file_text_proto_rawDescData
}

var file_text_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_text_proto_goTypes = []interface{}{
	(*CMIXChannelText)(nil),     // 0: parse.CMIXChannelText
	(*CMIXChannelReaction)(nil), // 1: parse.CMIXChannelReaction
}
var file_text_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_text_proto_init() }
func file_text_proto_init() {
	if File_text_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_text_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CMIXChannelText); i {
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
		file_text_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CMIXChannelReaction); i {
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
			RawDescriptor: file_text_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_text_proto_goTypes,
		DependencyIndexes: file_text_proto_depIdxs,
		MessageInfos:      file_text_proto_msgTypes,
	}.Build()
	File_text_proto = out.File
	file_text_proto_rawDesc = nil
	file_text_proto_goTypes = nil
	file_text_proto_depIdxs = nil
}