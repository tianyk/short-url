// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/short_url.proto

package proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
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

type ShortUrlMessage struct {
	LongUrl              string   `protobuf:"bytes,1,opt,name=longUrl,proto3" json:"longUrl,omitempty"`
	Expire               int64    `protobuf:"varint,2,opt,name=expire,proto3" json:"expire,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ShortUrlMessage) Reset()         { *m = ShortUrlMessage{} }
func (m *ShortUrlMessage) String() string { return proto.CompactTextString(m) }
func (*ShortUrlMessage) ProtoMessage()    {}
func (*ShortUrlMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_98c00b1926e3452e, []int{0}
}
func (m *ShortUrlMessage) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ShortUrlMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ShortUrlMessage.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ShortUrlMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ShortUrlMessage.Merge(m, src)
}
func (m *ShortUrlMessage) XXX_Size() int {
	return m.Size()
}
func (m *ShortUrlMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_ShortUrlMessage.DiscardUnknown(m)
}

var xxx_messageInfo_ShortUrlMessage proto.InternalMessageInfo

func (m *ShortUrlMessage) GetLongUrl() string {
	if m != nil {
		return m.LongUrl
	}
	return ""
}

func (m *ShortUrlMessage) GetExpire() int64 {
	if m != nil {
		return m.Expire
	}
	return 0
}

func init() {
	proto.RegisterType((*ShortUrlMessage)(nil), "ShortUrlMessage")
}

func init() { proto.RegisterFile("proto/short_url.proto", fileDescriptor_98c00b1926e3452e) }

var fileDescriptor_98c00b1926e3452e = []byte{
	// 127 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2d, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x2f, 0xce, 0xc8, 0x2f, 0x2a, 0x89, 0x2f, 0x2d, 0xca, 0xd1, 0x03, 0xf3, 0x95, 0x9c,
	0xb9, 0xf8, 0x83, 0x41, 0x42, 0xa1, 0x45, 0x39, 0xbe, 0xa9, 0xc5, 0xc5, 0x89, 0xe9, 0xa9, 0x42,
	0x12, 0x5c, 0xec, 0x39, 0xf9, 0x79, 0xe9, 0xa1, 0x45, 0x39, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c,
	0x41, 0x30, 0xae, 0x90, 0x18, 0x17, 0x5b, 0x6a, 0x45, 0x41, 0x66, 0x51, 0xaa, 0x04, 0x93, 0x02,
	0xa3, 0x06, 0x73, 0x10, 0x94, 0xe7, 0x24, 0x7e, 0xe2, 0x91, 0x1c, 0xe3, 0x85, 0x47, 0x72, 0x8c,
	0x0f, 0x1e, 0xc9, 0x31, 0xce, 0x78, 0x2c, 0xc7, 0x10, 0xc5, 0x0a, 0x36, 0x3d, 0x89, 0x0d, 0x4c,
	0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x87, 0xbc, 0x1e, 0xf9, 0x7d, 0x00, 0x00, 0x00,
}

func (m *ShortUrlMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ShortUrlMessage) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ShortUrlMessage) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Expire != 0 {
		i = encodeVarintShortUrl(dAtA, i, uint64(m.Expire))
		i--
		dAtA[i] = 0x10
	}
	if len(m.LongUrl) > 0 {
		i -= len(m.LongUrl)
		copy(dAtA[i:], m.LongUrl)
		i = encodeVarintShortUrl(dAtA, i, uint64(len(m.LongUrl)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintShortUrl(dAtA []byte, offset int, v uint64) int {
	offset -= sovShortUrl(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ShortUrlMessage) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.LongUrl)
	if l > 0 {
		n += 1 + l + sovShortUrl(uint64(l))
	}
	if m.Expire != 0 {
		n += 1 + sovShortUrl(uint64(m.Expire))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovShortUrl(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozShortUrl(x uint64) (n int) {
	return sovShortUrl(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ShortUrlMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowShortUrl
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ShortUrlMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ShortUrlMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LongUrl", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowShortUrl
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthShortUrl
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthShortUrl
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LongUrl = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Expire", wireType)
			}
			m.Expire = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowShortUrl
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Expire |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipShortUrl(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthShortUrl
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipShortUrl(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowShortUrl
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowShortUrl
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowShortUrl
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthShortUrl
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupShortUrl
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthShortUrl
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthShortUrl        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowShortUrl          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupShortUrl = fmt.Errorf("proto: unexpected end of group")
)
