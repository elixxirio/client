// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

/*
Package parse is a generated protocol buffer package.

It is generated from these files:
	types.proto

It has these top-level messages:
	TextMessage
	ChannelMessage
	PaymentInvoice
*/
package parse

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Type int32

const (
	Type_NO_TYPE         Type = 0
	Type_TEXT_MESSAGE    Type = 1
	Type_CHANNEL_MESSAGE Type = 2
	// We currently parse these types without using proto buffers
	// We use the types, but don't look for proto buffer definitions
	Type_UDB_PUSH_KEY          Type = 10
	Type_UDB_PUSH_KEY_RESPONSE Type = 11
	Type_UDB_GET_KEY           Type = 12
	Type_UDB_GET_KEY_RESPONSE  Type = 13
	Type_UDB_REGISTER          Type = 14
	Type_UDB_REGISTER_RESPONSE Type = 15
	Type_UDB_SEARCH            Type = 16
	Type_UDB_SEARCH_RESPONSE   Type = 17
	// Same with the payment bot types
	Type_PAYMENT Type = 20
	// Payment invoice uses a proto buffer because it might make things easier
	Type_PAYMENT_INVOICE Type = 21
)

var Type_name = map[int32]string{
	0:  "NO_TYPE",
	1:  "TEXT_MESSAGE",
	2:  "CHANNEL_MESSAGE",
	10: "UDB_PUSH_KEY",
	11: "UDB_PUSH_KEY_RESPONSE",
	12: "UDB_GET_KEY",
	13: "UDB_GET_KEY_RESPONSE",
	14: "UDB_REGISTER",
	15: "UDB_REGISTER_RESPONSE",
	16: "UDB_SEARCH",
	17: "UDB_SEARCH_RESPONSE",
	20: "PAYMENT",
	21: "PAYMENT_INVOICE",
}
var Type_value = map[string]int32{
	"NO_TYPE":               0,
	"TEXT_MESSAGE":          1,
	"CHANNEL_MESSAGE":       2,
	"UDB_PUSH_KEY":          10,
	"UDB_PUSH_KEY_RESPONSE": 11,
	"UDB_GET_KEY":           12,
	"UDB_GET_KEY_RESPONSE":  13,
	"UDB_REGISTER":          14,
	"UDB_REGISTER_RESPONSE": 15,
	"UDB_SEARCH":            16,
	"UDB_SEARCH_RESPONSE":   17,
	"PAYMENT":               20,
	"PAYMENT_INVOICE":       21,
}

func (x Type) String() string {
	return proto.EnumName(Type_name, int32(x))
}
func (Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type TextMessage struct {
	Color   int32  `protobuf:"zigzag32,2,opt,name=color" json:"color,omitempty"`
	Message string `protobuf:"bytes,3,opt,name=message" json:"message,omitempty"`
	Time    int64  `protobuf:"varint,4,opt,name=time" json:"time,omitempty"`
}

func (m *TextMessage) Reset()                    { *m = TextMessage{} }
func (m *TextMessage) String() string            { return proto.CompactTextString(m) }
func (*TextMessage) ProtoMessage()               {}
func (*TextMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *TextMessage) GetColor() int32 {
	if m != nil {
		return m.Color
	}
	return 0
}

func (m *TextMessage) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *TextMessage) GetTime() int64 {
	if m != nil {
		return m.Time
	}
	return 0
}

type ChannelMessage struct {
	SpeakerID []byte `protobuf:"bytes,3,opt,name=speakerID,proto3" json:"speakerID,omitempty"`
	Message   []byte `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
}

func (m *ChannelMessage) Reset()                    { *m = ChannelMessage{} }
func (m *ChannelMessage) String() string            { return proto.CompactTextString(m) }
func (*ChannelMessage) ProtoMessage()               {}
func (*ChannelMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ChannelMessage) GetSpeakerID() []byte {
	if m != nil {
		return m.SpeakerID
	}
	return nil
}

func (m *ChannelMessage) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

// Payment message types
type PaymentInvoice struct {
	Time         int64  `protobuf:"varint,1,opt,name=time" json:"time,omitempty"`
	CreatedCoins []byte `protobuf:"bytes,2,opt,name=createdCoins,proto3" json:"createdCoins,omitempty"`
	Memo         string `protobuf:"bytes,3,opt,name=memo" json:"memo,omitempty"`
}

func (m *PaymentInvoice) Reset()                    { *m = PaymentInvoice{} }
func (m *PaymentInvoice) String() string            { return proto.CompactTextString(m) }
func (*PaymentInvoice) ProtoMessage()               {}
func (*PaymentInvoice) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PaymentInvoice) GetTime() int64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *PaymentInvoice) GetCreatedCoins() []byte {
	if m != nil {
		return m.CreatedCoins
	}
	return nil
}

func (m *PaymentInvoice) GetMemo() string {
	if m != nil {
		return m.Memo
	}
	return ""
}

func init() {
	proto.RegisterType((*TextMessage)(nil), "parse.TextMessage")
	proto.RegisterType((*ChannelMessage)(nil), "parse.ChannelMessage")
	proto.RegisterType((*PaymentInvoice)(nil), "parse.PaymentInvoice")
	proto.RegisterEnum("parse.Type", Type_name, Type_value)
}

func init() { proto.RegisterFile("types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 364 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x92, 0xdf, 0x6e, 0xaa, 0x40,
	0x10, 0xc6, 0x0f, 0x8a, 0xc7, 0x38, 0x70, 0x70, 0x5d, 0x35, 0x87, 0x93, 0x9c, 0x0b, 0xc3, 0x95,
	0xe9, 0x45, 0x6f, 0xfa, 0x04, 0x88, 0x1b, 0x21, 0xad, 0x48, 0x97, 0xb5, 0xa9, 0x4d, 0x13, 0x42,
	0xed, 0xa4, 0x35, 0x95, 0x3f, 0x01, 0xd2, 0xd4, 0x57, 0xe8, 0x53, 0x37, 0xa0, 0x14, 0x7b, 0xb7,
	0xdf, 0x37, 0xbf, 0xfd, 0x66, 0x26, 0x19, 0x50, 0x8a, 0x43, 0x8a, 0xf9, 0x65, 0x9a, 0x25, 0x45,
	0x42, 0x3b, 0x69, 0x98, 0xe5, 0x68, 0xdc, 0x82, 0x22, 0xf0, 0xa3, 0x58, 0x62, 0x9e, 0x87, 0x2f,
	0x48, 0x47, 0xd0, 0xd9, 0x26, 0xfb, 0x24, 0xd3, 0x5b, 0x13, 0x69, 0x3a, 0xe0, 0x47, 0x41, 0x75,
	0xe8, 0x46, 0x47, 0x40, 0x6f, 0x4f, 0xa4, 0x69, 0x8f, 0xd7, 0x92, 0x52, 0x90, 0x8b, 0x5d, 0x84,
	0xba, 0x3c, 0x91, 0xa6, 0x6d, 0x5e, 0xbd, 0x0d, 0x1b, 0x34, 0xeb, 0x35, 0x8c, 0x63, 0xdc, 0xd7,
	0xa9, 0xff, 0xa1, 0x97, 0xa7, 0x18, 0xbe, 0x61, 0xe6, 0xcc, 0xab, 0x04, 0x95, 0x37, 0xc6, 0x79,
	0xba, 0x5c, 0xd5, 0x6a, 0x69, 0x3c, 0x82, 0xe6, 0x85, 0x87, 0x08, 0xe3, 0xc2, 0x89, 0xdf, 0x93,
	0xdd, 0xb6, 0xe9, 0x27, 0x35, 0xfd, 0xa8, 0x01, 0xea, 0x36, 0xc3, 0xb0, 0xc0, 0x67, 0x2b, 0xd9,
	0xc5, 0x79, 0x35, 0xba, 0xca, 0x7f, 0x78, 0xe5, 0xbf, 0x08, 0xa3, 0xe4, 0x34, 0x7e, 0xf5, 0xbe,
	0xf8, 0x6c, 0x81, 0x2c, 0x0e, 0x29, 0x52, 0x05, 0xba, 0xee, 0x2a, 0x10, 0x1b, 0x8f, 0x91, 0x5f,
	0x94, 0x80, 0x2a, 0xd8, 0xbd, 0x08, 0x96, 0xcc, 0xf7, 0xcd, 0x05, 0x23, 0x12, 0x1d, 0x42, 0xdf,
	0xb2, 0x4d, 0xd7, 0x65, 0x37, 0xdf, 0x66, 0xab, 0xc4, 0xd6, 0xf3, 0x59, 0xe0, 0xad, 0x7d, 0x3b,
	0xb8, 0x66, 0x1b, 0x02, 0xf4, 0x1f, 0x8c, 0xcf, 0x9d, 0x80, 0x33, 0xdf, 0x5b, 0xb9, 0x3e, 0x23,
	0x0a, 0xed, 0x83, 0x52, 0x96, 0x16, 0x4c, 0x54, 0xac, 0x4a, 0x75, 0x18, 0x9d, 0x19, 0x0d, 0xfa,
	0xa7, 0xce, 0xe5, 0x6c, 0xe1, 0xf8, 0x82, 0x71, 0xa2, 0xd5, 0xb9, 0xb5, 0xd3, 0xc0, 0x7d, 0xaa,
	0x01, 0x94, 0x25, 0x9f, 0x99, 0xdc, 0xb2, 0x09, 0xa1, 0x7f, 0x61, 0xd8, 0xe8, 0x06, 0x1c, 0x94,
	0x1b, 0x7a, 0xe6, 0x66, 0xc9, 0x5c, 0x41, 0x46, 0xe5, 0x3e, 0x27, 0x11, 0x38, 0xee, 0xdd, 0xca,
	0xb1, 0x18, 0x19, 0xcf, 0xba, 0x0f, 0xc7, 0x83, 0x78, 0xfa, 0x5d, 0x9d, 0xc7, 0xd5, 0x57, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x83, 0xac, 0x72, 0x30, 0x2d, 0x02, 0x00, 0x00,
}
