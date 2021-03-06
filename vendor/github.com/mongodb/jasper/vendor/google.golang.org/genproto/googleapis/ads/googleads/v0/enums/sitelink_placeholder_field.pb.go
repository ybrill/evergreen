// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v0/enums/sitelink_placeholder_field.proto

package enums // import "google.golang.org/genproto/googleapis/ads/googleads/v0/enums"

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

// Possible values for Sitelink placeholder fields.
type SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField int32

const (
	// Not specified.
	SitelinkPlaceholderFieldEnum_UNSPECIFIED SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 0
	// Used for return value only. Represents value unknown in this version.
	SitelinkPlaceholderFieldEnum_UNKNOWN SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 1
	// Data Type: STRING. The link text for your sitelink.
	SitelinkPlaceholderFieldEnum_TEXT SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 2
	// Data Type: STRING. First line of the sitelink description.
	SitelinkPlaceholderFieldEnum_LINE_1 SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 3
	// Data Type: STRING. Second line of the sitelink description.
	SitelinkPlaceholderFieldEnum_LINE_2 SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 4
	// Data Type: URL_LIST. Final URLs for the sitelink when using Upgraded
	// URLs.
	SitelinkPlaceholderFieldEnum_FINAL_URLS SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 5
	// Data Type: URL_LIST. Final Mobile URLs for the sitelink when using
	// Upgraded URLs.
	SitelinkPlaceholderFieldEnum_FINAL_MOBILE_URLS SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 6
	// Data Type: URL. Tracking template for the sitelink when using Upgraded
	// URLs.
	SitelinkPlaceholderFieldEnum_TRACKING_URL SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 7
	// Data Type: STRING. Final URL suffix for sitelink when using parallel
	// tracking.
	SitelinkPlaceholderFieldEnum_FINAL_URL_SUFFIX SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField = 8
)

var SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UNKNOWN",
	2: "TEXT",
	3: "LINE_1",
	4: "LINE_2",
	5: "FINAL_URLS",
	6: "FINAL_MOBILE_URLS",
	7: "TRACKING_URL",
	8: "FINAL_URL_SUFFIX",
}
var SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField_value = map[string]int32{
	"UNSPECIFIED":       0,
	"UNKNOWN":           1,
	"TEXT":              2,
	"LINE_1":            3,
	"LINE_2":            4,
	"FINAL_URLS":        5,
	"FINAL_MOBILE_URLS": 6,
	"TRACKING_URL":      7,
	"FINAL_URL_SUFFIX":  8,
}

func (x SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField) String() string {
	return proto.EnumName(SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField_name, int32(x))
}
func (SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_sitelink_placeholder_field_37144d784e8143c8, []int{0, 0}
}

// Values for Sitelink placeholder fields.
type SitelinkPlaceholderFieldEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SitelinkPlaceholderFieldEnum) Reset()         { *m = SitelinkPlaceholderFieldEnum{} }
func (m *SitelinkPlaceholderFieldEnum) String() string { return proto.CompactTextString(m) }
func (*SitelinkPlaceholderFieldEnum) ProtoMessage()    {}
func (*SitelinkPlaceholderFieldEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_sitelink_placeholder_field_37144d784e8143c8, []int{0}
}
func (m *SitelinkPlaceholderFieldEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SitelinkPlaceholderFieldEnum.Unmarshal(m, b)
}
func (m *SitelinkPlaceholderFieldEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SitelinkPlaceholderFieldEnum.Marshal(b, m, deterministic)
}
func (dst *SitelinkPlaceholderFieldEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SitelinkPlaceholderFieldEnum.Merge(dst, src)
}
func (m *SitelinkPlaceholderFieldEnum) XXX_Size() int {
	return xxx_messageInfo_SitelinkPlaceholderFieldEnum.Size(m)
}
func (m *SitelinkPlaceholderFieldEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_SitelinkPlaceholderFieldEnum.DiscardUnknown(m)
}

var xxx_messageInfo_SitelinkPlaceholderFieldEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*SitelinkPlaceholderFieldEnum)(nil), "google.ads.googleads.v0.enums.SitelinkPlaceholderFieldEnum")
	proto.RegisterEnum("google.ads.googleads.v0.enums.SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField", SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField_name, SitelinkPlaceholderFieldEnum_SitelinkPlaceholderField_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v0/enums/sitelink_placeholder_field.proto", fileDescriptor_sitelink_placeholder_field_37144d784e8143c8)
}

var fileDescriptor_sitelink_placeholder_field_37144d784e8143c8 = []byte{
	// 357 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x51, 0xcd, 0x6e, 0xaa, 0x40,
	0x14, 0xbe, 0xa0, 0x57, 0xcd, 0xf1, 0xe6, 0xde, 0xb9, 0x93, 0x36, 0xe9, 0xa2, 0x2e, 0xf4, 0x01,
	0x06, 0xda, 0xee, 0xa6, 0x49, 0x13, 0xb0, 0x60, 0x88, 0x14, 0x89, 0x88, 0x35, 0x0d, 0x09, 0xa1,
	0x42, 0x29, 0x29, 0x32, 0xc6, 0x51, 0x1f, 0xa8, 0xbb, 0xf6, 0x51, 0x7c, 0x94, 0x2e, 0xfb, 0x04,
	0x0d, 0xa0, 0x74, 0x65, 0x37, 0xe4, 0xe3, 0x7c, 0x3f, 0x99, 0xf3, 0x1d, 0xb8, 0x89, 0x19, 0x8b,
	0xd3, 0x48, 0x0a, 0x42, 0x2e, 0x95, 0x30, 0x47, 0x5b, 0x59, 0x8a, 0xb2, 0xcd, 0x82, 0x4b, 0x3c,
	0x59, 0x47, 0x69, 0x92, 0xbd, 0xf8, 0xcb, 0x34, 0x98, 0x47, 0xcf, 0x2c, 0x0d, 0xa3, 0x95, 0xff,
	0x94, 0x44, 0x69, 0x48, 0x96, 0x2b, 0xb6, 0x66, 0xb8, 0x53, 0x9a, 0x48, 0x10, 0x72, 0x52, 0xf9,
	0xc9, 0x56, 0x26, 0x85, 0xbf, 0xb7, 0x13, 0xe0, 0xdc, 0xd9, 0x67, 0xd8, 0xdf, 0x11, 0x7a, 0x9e,
	0xa0, 0x65, 0x9b, 0x45, 0xef, 0x4d, 0x80, 0xb3, 0x63, 0x02, 0xfc, 0x0f, 0xda, 0xae, 0xe5, 0xd8,
	0x5a, 0xdf, 0xd0, 0x0d, 0xed, 0x16, 0xfd, 0xc2, 0x6d, 0x68, 0xba, 0xd6, 0xd0, 0x1a, 0xdd, 0x5b,
	0x48, 0xc0, 0x2d, 0xa8, 0x4f, 0xb4, 0xd9, 0x04, 0x89, 0x18, 0xa0, 0x61, 0x1a, 0x96, 0xe6, 0x5f,
	0xa0, 0x5a, 0x85, 0x2f, 0x51, 0x1d, 0xff, 0x05, 0xd0, 0x0d, 0x4b, 0x31, 0x7d, 0x77, 0x6c, 0x3a,
	0xe8, 0x37, 0x3e, 0x85, 0xff, 0xe5, 0xff, 0xdd, 0x48, 0x35, 0x4c, 0xad, 0x1c, 0x37, 0x30, 0x82,
	0x3f, 0x93, 0xb1, 0xd2, 0x1f, 0x1a, 0xd6, 0x20, 0x1f, 0xa1, 0x26, 0x3e, 0x01, 0x54, 0x19, 0x7d,
	0xc7, 0xd5, 0x75, 0x63, 0x86, 0x5a, 0xea, 0xa7, 0x00, 0xdd, 0x39, 0x5b, 0x90, 0x1f, 0x57, 0x56,
	0x3b, 0xc7, 0xd6, 0xb1, 0xf3, 0xc2, 0x6c, 0xe1, 0x41, 0xdd, 0xfb, 0x63, 0x96, 0x06, 0x59, 0x4c,
	0xd8, 0x2a, 0x96, 0xe2, 0x28, 0x2b, 0xea, 0x3c, 0x9c, 0x60, 0x99, 0xf0, 0x23, 0x17, 0xb9, 0x2e,
	0xbe, 0xaf, 0x62, 0x6d, 0xa0, 0x28, 0xef, 0x62, 0x67, 0x50, 0x46, 0x29, 0x21, 0x27, 0x25, 0xcc,
	0xd1, 0x54, 0x26, 0x79, 0xb7, 0x7c, 0x77, 0xe0, 0x3d, 0x25, 0xe4, 0x5e, 0xc5, 0x7b, 0x53, 0xd9,
	0x2b, 0xf8, 0x0f, 0xb1, 0x5b, 0x0e, 0x29, 0x55, 0x42, 0x4e, 0x69, 0xa5, 0xa0, 0x74, 0x2a, 0x53,
	0x5a, 0x68, 0x1e, 0x1b, 0xc5, 0xc3, 0xae, 0xbe, 0x02, 0x00, 0x00, 0xff, 0xff, 0x37, 0x1b, 0xd9,
	0xbf, 0x29, 0x02, 0x00, 0x00,
}
