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
// source: udMessages.proto

package ud

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

// Contains the Hash and its Type
type HashFact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Type int32  `protobuf:"varint,2,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *HashFact) Reset() {
	*x = HashFact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_udMessages_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HashFact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HashFact) ProtoMessage() {}

func (x *HashFact) ProtoReflect() protoreflect.Message {
	mi := &file_udMessages_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HashFact.ProtoReflect.Descriptor instead.
func (*HashFact) Descriptor() ([]byte, []int) {
	return file_udMessages_proto_rawDescGZIP(), []int{0}
}

func (x *HashFact) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *HashFact) GetType() int32 {
	if x != nil {
		return x.Type
	}
	return 0
}

// Describes a user lookup result. The ID, public key, and the
// facts inputted that brought up this user.
type Contact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID    []byte      `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	PubKey    []byte      `protobuf:"bytes,2,opt,name=pubKey,proto3" json:"pubKey,omitempty"`
	Username  string      `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
	TrigFacts []*HashFact `protobuf:"bytes,4,rep,name=trigFacts,proto3" json:"trigFacts,omitempty"`
}

func (x *Contact) Reset() {
	*x = Contact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_udMessages_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Contact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Contact) ProtoMessage() {}

func (x *Contact) ProtoReflect() protoreflect.Message {
	mi := &file_udMessages_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Contact.ProtoReflect.Descriptor instead.
func (*Contact) Descriptor() ([]byte, []int) {
	return file_udMessages_proto_rawDescGZIP(), []int{1}
}

func (x *Contact) GetUserID() []byte {
	if x != nil {
		return x.UserID
	}
	return nil
}

func (x *Contact) GetPubKey() []byte {
	if x != nil {
		return x.PubKey
	}
	return nil
}

func (x *Contact) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Contact) GetTrigFacts() []*HashFact {
	if x != nil {
		return x.TrigFacts
	}
	return nil
}

// Message sent to UDB to search for users
type SearchSend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PublicKey used in the registration
	Fact []*HashFact `protobuf:"bytes,1,rep,name=fact,proto3" json:"fact,omitempty"`
}

func (x *SearchSend) Reset() {
	*x = SearchSend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_udMessages_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchSend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchSend) ProtoMessage() {}

func (x *SearchSend) ProtoReflect() protoreflect.Message {
	mi := &file_udMessages_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchSend.ProtoReflect.Descriptor instead.
func (*SearchSend) Descriptor() ([]byte, []int) {
	return file_udMessages_proto_rawDescGZIP(), []int{2}
}

func (x *SearchSend) GetFact() []*HashFact {
	if x != nil {
		return x.Fact
	}
	return nil
}

// Message sent from UDB to client in response to a search
type SearchResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID of the session created
	Contacts []*Contact `protobuf:"bytes,1,rep,name=contacts,proto3" json:"contacts,omitempty"`
	Error    string     `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *SearchResponse) Reset() {
	*x = SearchResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_udMessages_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchResponse) ProtoMessage() {}

func (x *SearchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_udMessages_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchResponse.ProtoReflect.Descriptor instead.
func (*SearchResponse) Descriptor() ([]byte, []int) {
	return file_udMessages_proto_rawDescGZIP(), []int{3}
}

func (x *SearchResponse) GetContacts() []*Contact {
	if x != nil {
		return x.Contacts
	}
	return nil
}

func (x *SearchResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

// Message sent to UDB for looking up a user
type LookupSend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID []byte `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
}

func (x *LookupSend) Reset() {
	*x = LookupSend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_udMessages_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupSend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupSend) ProtoMessage() {}

func (x *LookupSend) ProtoReflect() protoreflect.Message {
	mi := &file_udMessages_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupSend.ProtoReflect.Descriptor instead.
func (*LookupSend) Descriptor() ([]byte, []int) {
	return file_udMessages_proto_rawDescGZIP(), []int{4}
}

func (x *LookupSend) GetUserID() []byte {
	if x != nil {
		return x.UserID
	}
	return nil
}

// Message sent from UDB for looking up a user
type LookupResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PubKey   []byte `protobuf:"bytes,1,opt,name=pubKey,proto3" json:"pubKey,omitempty"`
	Username string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Error    string `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *LookupResponse) Reset() {
	*x = LookupResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_udMessages_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupResponse) ProtoMessage() {}

func (x *LookupResponse) ProtoReflect() protoreflect.Message {
	mi := &file_udMessages_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupResponse.ProtoReflect.Descriptor instead.
func (*LookupResponse) Descriptor() ([]byte, []int) {
	return file_udMessages_proto_rawDescGZIP(), []int{5}
}

func (x *LookupResponse) GetPubKey() []byte {
	if x != nil {
		return x.PubKey
	}
	return nil
}

func (x *LookupResponse) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *LookupResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_udMessages_proto protoreflect.FileDescriptor

var file_udMessages_proto_rawDesc = []byte{
	0x0a, 0x10, 0x75, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x02, 0x75, 0x64, 0x22, 0x32, 0x0a, 0x08, 0x48, 0x61, 0x73, 0x68, 0x46, 0x61,
	0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x81, 0x01, 0x0a, 0x07, 0x43,
	0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x16,
	0x0a, 0x06, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06,
	0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x2a, 0x0a, 0x09, 0x74, 0x72, 0x69, 0x67, 0x46, 0x61, 0x63, 0x74, 0x73, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x75, 0x64, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x46,
	0x61, 0x63, 0x74, 0x52, 0x09, 0x74, 0x72, 0x69, 0x67, 0x46, 0x61, 0x63, 0x74, 0x73, 0x22, 0x2e,
	0x0a, 0x0a, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x53, 0x65, 0x6e, 0x64, 0x12, 0x20, 0x0a, 0x04,
	0x66, 0x61, 0x63, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x75, 0x64, 0x2e,
	0x48, 0x61, 0x73, 0x68, 0x46, 0x61, 0x63, 0x74, 0x52, 0x04, 0x66, 0x61, 0x63, 0x74, 0x22, 0x4f,
	0x0a, 0x0e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x27, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x75, 0x64, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x52,
	0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22,
	0x24, 0x0a, 0x0a, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x53, 0x65, 0x6e, 0x64, 0x12, 0x16, 0x0a,
	0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x75,
	0x73, 0x65, 0x72, 0x49, 0x44, 0x22, 0x5a, 0x0a, 0x0e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x75, 0x62, 0x4b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x12,
	0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x42, 0x1e, 0x5a, 0x1c, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x65, 0x6c, 0x69, 0x78, 0x78, 0x69, 0x72, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x75,
	0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_udMessages_proto_rawDescOnce sync.Once
	file_udMessages_proto_rawDescData = file_udMessages_proto_rawDesc
)

func file_udMessages_proto_rawDescGZIP() []byte {
	file_udMessages_proto_rawDescOnce.Do(func() {
		file_udMessages_proto_rawDescData = protoimpl.X.CompressGZIP(file_udMessages_proto_rawDescData)
	})
	return file_udMessages_proto_rawDescData
}

var file_udMessages_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_udMessages_proto_goTypes = []interface{}{
	(*HashFact)(nil),       // 0: ud.HashFact
	(*Contact)(nil),        // 1: ud.Contact
	(*SearchSend)(nil),     // 2: ud.SearchSend
	(*SearchResponse)(nil), // 3: ud.SearchResponse
	(*LookupSend)(nil),     // 4: ud.LookupSend
	(*LookupResponse)(nil), // 5: ud.LookupResponse
}
var file_udMessages_proto_depIdxs = []int32{
	0, // 0: ud.Contact.trigFacts:type_name -> ud.HashFact
	0, // 1: ud.SearchSend.fact:type_name -> ud.HashFact
	1, // 2: ud.SearchResponse.contacts:type_name -> ud.Contact
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_udMessages_proto_init() }
func file_udMessages_proto_init() {
	if File_udMessages_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_udMessages_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HashFact); i {
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
		file_udMessages_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Contact); i {
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
		file_udMessages_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchSend); i {
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
		file_udMessages_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchResponse); i {
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
		file_udMessages_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupSend); i {
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
		file_udMessages_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupResponse); i {
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
			RawDescriptor: file_udMessages_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_udMessages_proto_goTypes,
		DependencyIndexes: file_udMessages_proto_depIdxs,
		MessageInfos:      file_udMessages_proto_msgTypes,
	}.Build()
	File_udMessages_proto = out.File
	file_udMessages_proto_rawDesc = nil
	file_udMessages_proto_goTypes = nil
	file_udMessages_proto_depIdxs = nil
}
