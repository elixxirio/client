///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.15.6
// source: authenticated.proto

package connect

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

// Sent by the receiver of the authenticated connection request.
type IdentityAuthentication struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Signature []byte `protobuf:"bytes,1,opt,name=Signature,proto3" json:"Signature,omitempty"` // Signature of the connection fingerprint
	// established between the two partners
	RsaPubKey []byte `protobuf:"bytes,2,opt,name=RsaPubKey,proto3" json:"RsaPubKey,omitempty"` // The RSA public key of the sender of this message,
	// PEM-encoded
	Salt []byte `protobuf:"bytes,3,opt,name=Salt,proto3" json:"Salt,omitempty"` // Salt used to generate the network ID of the client
}

func (x *IdentityAuthentication) Reset() {
	*x = IdentityAuthentication{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authenticated_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IdentityAuthentication) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IdentityAuthentication) ProtoMessage() {}

func (x *IdentityAuthentication) ProtoReflect() protoreflect.Message {
	mi := &file_authenticated_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IdentityAuthentication.ProtoReflect.Descriptor instead.
func (*IdentityAuthentication) Descriptor() ([]byte, []int) {
	return file_authenticated_proto_rawDescGZIP(), []int{0}
}

func (x *IdentityAuthentication) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *IdentityAuthentication) GetRsaPubKey() []byte {
	if x != nil {
		return x.RsaPubKey
	}
	return nil
}

func (x *IdentityAuthentication) GetSalt() []byte {
	if x != nil {
		return x.Salt
	}
	return nil
}

var File_authenticated_proto protoreflect.FileDescriptor

var file_authenticated_proto_rawDesc = []byte{
	0x0a, 0x13, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x22, 0x68,
	0x0a, 0x16, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e,
	0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x53, 0x69, 0x67,
	0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x52, 0x73, 0x61, 0x50, 0x75, 0x62,
	0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x52, 0x73, 0x61, 0x50, 0x75,
	0x62, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x53, 0x61, 0x6c, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x53, 0x61, 0x6c, 0x74, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69, 0x74, 0x6c,
	0x61, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6c, 0x69, 0x78, 0x78, 0x69, 0x72, 0x2f, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_authenticated_proto_rawDescOnce sync.Once
	file_authenticated_proto_rawDescData = file_authenticated_proto_rawDesc
)

func file_authenticated_proto_rawDescGZIP() []byte {
	file_authenticated_proto_rawDescOnce.Do(func() {
		file_authenticated_proto_rawDescData = protoimpl.X.CompressGZIP(file_authenticated_proto_rawDescData)
	})
	return file_authenticated_proto_rawDescData
}

var file_authenticated_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_authenticated_proto_goTypes = []interface{}{
	(*IdentityAuthentication)(nil), // 0: connect.IdentityAuthentication
}
var file_authenticated_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_authenticated_proto_init() }
func file_authenticated_proto_init() {
	if File_authenticated_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_authenticated_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IdentityAuthentication); i {
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
			RawDescriptor: file_authenticated_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_authenticated_proto_goTypes,
		DependencyIndexes: file_authenticated_proto_depIdxs,
		MessageInfos:      file_authenticated_proto_msgTypes,
	}.Build()
	File_authenticated_proto = out.File
	file_authenticated_proto_rawDesc = nil
	file_authenticated_proto_goTypes = nil
	file_authenticated_proto_depIdxs = nil
}
