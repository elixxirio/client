// Code generated by protoc-gen-go. DO NOT EDIT.
// source: udMessages.proto

package ud

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Contains the Hash and its Type
type HashFact struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Type                 int32    `protobuf:"varint,2,opt,name=type,proto3" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HashFact) Reset()         { *m = HashFact{} }
func (m *HashFact) String() string { return proto.CompactTextString(m) }
func (*HashFact) ProtoMessage()    {}
func (*HashFact) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e0cfdc16fb09bb6, []int{0}
}

func (m *HashFact) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HashFact.Unmarshal(m, b)
}
func (m *HashFact) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HashFact.Marshal(b, m, deterministic)
}
func (m *HashFact) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HashFact.Merge(m, src)
}
func (m *HashFact) XXX_Size() int {
	return xxx_messageInfo_HashFact.Size(m)
}
func (m *HashFact) XXX_DiscardUnknown() {
	xxx_messageInfo_HashFact.DiscardUnknown(m)
}

var xxx_messageInfo_HashFact proto.InternalMessageInfo

func (m *HashFact) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *HashFact) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}

// Describes a user lookup result. The ID, public key, and the
// facts inputted that brought up this user.
type Contact struct {
	UserID               []byte      `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	PubKey               []byte      `protobuf:"bytes,2,opt,name=pubKey,proto3" json:"pubKey,omitempty"`
	TrigFacts            []*HashFact `protobuf:"bytes,3,rep,name=trigFacts,proto3" json:"trigFacts,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Contact) Reset()         { *m = Contact{} }
func (m *Contact) String() string { return proto.CompactTextString(m) }
func (*Contact) ProtoMessage()    {}
func (*Contact) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e0cfdc16fb09bb6, []int{1}
}

func (m *Contact) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Contact.Unmarshal(m, b)
}
func (m *Contact) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Contact.Marshal(b, m, deterministic)
}
func (m *Contact) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Contact.Merge(m, src)
}
func (m *Contact) XXX_Size() int {
	return xxx_messageInfo_Contact.Size(m)
}
func (m *Contact) XXX_DiscardUnknown() {
	xxx_messageInfo_Contact.DiscardUnknown(m)
}

var xxx_messageInfo_Contact proto.InternalMessageInfo

func (m *Contact) GetUserID() []byte {
	if m != nil {
		return m.UserID
	}
	return nil
}

func (m *Contact) GetPubKey() []byte {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *Contact) GetTrigFacts() []*HashFact {
	if m != nil {
		return m.TrigFacts
	}
	return nil
}

// Message sent to UDB to search for users
type SearchSend struct {
	// PublicKey used in the registration
	Fact []*HashFact `protobuf:"bytes,1,rep,name=fact,proto3" json:"fact,omitempty"`
	// ID of the session used to create this session
	CommID               uint64   `protobuf:"varint,2,opt,name=commID,proto3" json:"commID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SearchSend) Reset()         { *m = SearchSend{} }
func (m *SearchSend) String() string { return proto.CompactTextString(m) }
func (*SearchSend) ProtoMessage()    {}
func (*SearchSend) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e0cfdc16fb09bb6, []int{2}
}

func (m *SearchSend) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SearchSend.Unmarshal(m, b)
}
func (m *SearchSend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SearchSend.Marshal(b, m, deterministic)
}
func (m *SearchSend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SearchSend.Merge(m, src)
}
func (m *SearchSend) XXX_Size() int {
	return xxx_messageInfo_SearchSend.Size(m)
}
func (m *SearchSend) XXX_DiscardUnknown() {
	xxx_messageInfo_SearchSend.DiscardUnknown(m)
}

var xxx_messageInfo_SearchSend proto.InternalMessageInfo

func (m *SearchSend) GetFact() []*HashFact {
	if m != nil {
		return m.Fact
	}
	return nil
}

func (m *SearchSend) GetCommID() uint64 {
	if m != nil {
		return m.CommID
	}
	return 0
}

// Message sent from UDB to client in response to a search
type SearchResponse struct {
	// ID of the session created
	Contacts             []*Contact `protobuf:"bytes,1,rep,name=contacts,proto3" json:"contacts,omitempty"`
	CommID               uint64     `protobuf:"varint,2,opt,name=commID,proto3" json:"commID,omitempty"`
	Error                string     `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *SearchResponse) Reset()         { *m = SearchResponse{} }
func (m *SearchResponse) String() string { return proto.CompactTextString(m) }
func (*SearchResponse) ProtoMessage()    {}
func (*SearchResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e0cfdc16fb09bb6, []int{3}
}

func (m *SearchResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SearchResponse.Unmarshal(m, b)
}
func (m *SearchResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SearchResponse.Marshal(b, m, deterministic)
}
func (m *SearchResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SearchResponse.Merge(m, src)
}
func (m *SearchResponse) XXX_Size() int {
	return xxx_messageInfo_SearchResponse.Size(m)
}
func (m *SearchResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SearchResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SearchResponse proto.InternalMessageInfo

func (m *SearchResponse) GetContacts() []*Contact {
	if m != nil {
		return m.Contacts
	}
	return nil
}

func (m *SearchResponse) GetCommID() uint64 {
	if m != nil {
		return m.CommID
	}
	return 0
}

func (m *SearchResponse) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

// Message sent to UDB for looking up a user
type LookupSend struct {
	UserID               []byte   `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	CommID               uint64   `protobuf:"varint,2,opt,name=commID,proto3" json:"commID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LookupSend) Reset()         { *m = LookupSend{} }
func (m *LookupSend) String() string { return proto.CompactTextString(m) }
func (*LookupSend) ProtoMessage()    {}
func (*LookupSend) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e0cfdc16fb09bb6, []int{4}
}

func (m *LookupSend) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LookupSend.Unmarshal(m, b)
}
func (m *LookupSend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LookupSend.Marshal(b, m, deterministic)
}
func (m *LookupSend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LookupSend.Merge(m, src)
}
func (m *LookupSend) XXX_Size() int {
	return xxx_messageInfo_LookupSend.Size(m)
}
func (m *LookupSend) XXX_DiscardUnknown() {
	xxx_messageInfo_LookupSend.DiscardUnknown(m)
}

var xxx_messageInfo_LookupSend proto.InternalMessageInfo

func (m *LookupSend) GetUserID() []byte {
	if m != nil {
		return m.UserID
	}
	return nil
}

func (m *LookupSend) GetCommID() uint64 {
	if m != nil {
		return m.CommID
	}
	return 0
}

// Message sent from UDB for looking up a user
type LookupResponse struct {
	PubKey               []byte   `protobuf:"bytes,1,opt,name=pubKey,proto3" json:"pubKey,omitempty"`
	CommID               uint64   `protobuf:"varint,2,opt,name=commID,proto3" json:"commID,omitempty"`
	Error                string   `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LookupResponse) Reset()         { *m = LookupResponse{} }
func (m *LookupResponse) String() string { return proto.CompactTextString(m) }
func (*LookupResponse) ProtoMessage()    {}
func (*LookupResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e0cfdc16fb09bb6, []int{5}
}

func (m *LookupResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LookupResponse.Unmarshal(m, b)
}
func (m *LookupResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LookupResponse.Marshal(b, m, deterministic)
}
func (m *LookupResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LookupResponse.Merge(m, src)
}
func (m *LookupResponse) XXX_Size() int {
	return xxx_messageInfo_LookupResponse.Size(m)
}
func (m *LookupResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_LookupResponse.DiscardUnknown(m)
}

var xxx_messageInfo_LookupResponse proto.InternalMessageInfo

func (m *LookupResponse) GetPubKey() []byte {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *LookupResponse) GetCommID() uint64 {
	if m != nil {
		return m.CommID
	}
	return 0
}

func (m *LookupResponse) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*HashFact)(nil), "parse.HashFact")
	proto.RegisterType((*Contact)(nil), "parse.Contact")
	proto.RegisterType((*SearchSend)(nil), "parse.SearchSend")
	proto.RegisterType((*SearchResponse)(nil), "parse.SearchResponse")
	proto.RegisterType((*LookupSend)(nil), "parse.LookupSend")
	proto.RegisterType((*LookupResponse)(nil), "parse.LookupResponse")
}

func init() {
	proto.RegisterFile("udMessages.proto", fileDescriptor_9e0cfdc16fb09bb6)
}

var fileDescriptor_9e0cfdc16fb09bb6 = []byte{
	// 285 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x51, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x65, 0x9b, 0xa4, 0xb6, 0x63, 0x89, 0xb2, 0x88, 0xe4, 0x18, 0xe2, 0x25, 0x08, 0xe6, 0x50,
	0xaf, 0x9e, 0xb4, 0x88, 0x41, 0xbd, 0x6c, 0xc1, 0x83, 0xb7, 0x6d, 0x32, 0x36, 0x2a, 0xcd, 0x2e,
	0x3b, 0x9b, 0x43, 0xff, 0x5e, 0xb2, 0x59, 0x5b, 0x84, 0x2a, 0x78, 0x9b, 0x37, 0xb3, 0xef, 0xcd,
	0x9b, 0xb7, 0x70, 0xda, 0xd5, 0xcf, 0x48, 0x24, 0xd7, 0x48, 0x85, 0x36, 0xca, 0x2a, 0x1e, 0x69,
	0x69, 0x08, 0xb3, 0x39, 0x4c, 0x1e, 0x24, 0x35, 0xf7, 0xb2, 0xb2, 0x9c, 0x43, 0xd8, 0x48, 0x6a,
	0x12, 0x96, 0xb2, 0x7c, 0x26, 0x5c, 0xdd, 0xf7, 0xec, 0x56, 0x63, 0x32, 0x4a, 0x59, 0x1e, 0x09,
	0x57, 0x67, 0x0d, 0x1c, 0xdd, 0xa9, 0xd6, 0xf6, 0x94, 0x73, 0x18, 0x77, 0x84, 0xa6, 0x5c, 0x78,
	0x92, 0x47, 0x7d, 0x5f, 0x77, 0xab, 0x47, 0xdc, 0x3a, 0xe2, 0x4c, 0x78, 0xc4, 0xaf, 0x60, 0x6a,
	0xcd, 0xfb, 0xba, 0x5f, 0x47, 0x49, 0x90, 0x06, 0xf9, 0xf1, 0xfc, 0xa4, 0x70, 0x4e, 0x8a, 0x6f,
	0x1b, 0x62, 0xff, 0x22, 0x2b, 0x01, 0x96, 0x28, 0x4d, 0xd5, 0x2c, 0xb1, 0xad, 0xf9, 0x05, 0x84,
	0x6f, 0xb2, 0xb2, 0x09, 0x3b, 0xcc, 0x73, 0xc3, 0x7e, 0x73, 0xa5, 0x36, 0x9b, 0x72, 0xe1, 0x36,
	0x87, 0xc2, 0xa3, 0xec, 0x03, 0xe2, 0x41, 0x4a, 0x20, 0x69, 0xd5, 0x12, 0xf2, 0x4b, 0x98, 0x54,
	0xc3, 0x19, 0xe4, 0x25, 0x63, 0x2f, 0xe9, 0xaf, 0x13, 0xbb, 0xf9, 0x6f, 0xaa, 0xfc, 0x0c, 0x22,
	0x34, 0x46, 0x99, 0x24, 0x48, 0x59, 0x3e, 0x15, 0x03, 0xc8, 0x6e, 0x00, 0x9e, 0x94, 0xfa, 0xec,
	0xb4, 0xb3, 0xfd, 0x47, 0x46, 0x07, 0x9d, 0xbe, 0x40, 0x3c, 0xb0, 0x77, 0x4e, 0xf7, 0x69, 0xb2,
	0x1f, 0x69, 0xfe, 0xcb, 0xd5, 0x6d, 0xf8, 0x3a, 0xea, 0xea, 0xd5, 0xd8, 0x7d, 0xff, 0xf5, 0x57,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x3d, 0x49, 0xa1, 0x5e, 0x12, 0x02, 0x00, 0x00,
}
