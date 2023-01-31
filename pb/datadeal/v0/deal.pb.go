// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.6.1
// source: panacea/datadeal/v0/deal.proto

package v0

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type ValidateDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DealId          uint64 `protobuf:"varint,1,opt,name=deal_id,proto3" json:"deal_id,omitempty"`
	ProviderAddress string `protobuf:"bytes,2,opt,name=provider_address,proto3" json:"provider_address,omitempty"`
	EncryptedData   []byte `protobuf:"bytes,3,opt,name=encrypted_data,proto3" json:"encrypted_data,omitempty"`
	DataHash        []byte `protobuf:"bytes,4,opt,name=data_hash,proto3" json:"data_hash,omitempty"`
}

func (x *ValidateDataRequest) Reset() {
	*x = ValidateDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateDataRequest) ProtoMessage() {}

func (x *ValidateDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateDataRequest.ProtoReflect.Descriptor instead.
func (*ValidateDataRequest) Descriptor() ([]byte, []int) {
	return file_panacea_datadeal_v0_deal_proto_rawDescGZIP(), []int{0}
}

func (x *ValidateDataRequest) GetDealId() uint64 {
	if x != nil {
		return x.DealId
	}
	return 0
}

func (x *ValidateDataRequest) GetProviderAddress() string {
	if x != nil {
		return x.ProviderAddress
	}
	return ""
}

func (x *ValidateDataRequest) GetEncryptedData() []byte {
	if x != nil {
		return x.EncryptedData
	}
	return nil
}

func (x *ValidateDataRequest) GetDataHash() []byte {
	if x != nil {
		return x.DataHash
	}
	return nil
}

type ValidateDataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Certificate *Certificate `protobuf:"bytes,1,opt,name=certificate,proto3" json:"certificate,omitempty"`
}

func (x *ValidateDataResponse) Reset() {
	*x = ValidateDataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateDataResponse) ProtoMessage() {}

func (x *ValidateDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateDataResponse.ProtoReflect.Descriptor instead.
func (*ValidateDataResponse) Descriptor() ([]byte, []int) {
	return file_panacea_datadeal_v0_deal_proto_rawDescGZIP(), []int{1}
}

func (x *ValidateDataResponse) GetCertificate() *Certificate {
	if x != nil {
		return x.Certificate
	}
	return nil
}

// Certificate defines a certificate signed by an oracle who issued the certificate.
type Certificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UnsignedCertificate *UnsignedCertificate `protobuf:"bytes,1,opt,name=unsigned_certificate,proto3" json:"unsigned_certificate,omitempty"`
	Signature           []byte               `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *Certificate) Reset() {
	*x = Certificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Certificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Certificate) ProtoMessage() {}

func (x *Certificate) ProtoReflect() protoreflect.Message {
	mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Certificate.ProtoReflect.Descriptor instead.
func (*Certificate) Descriptor() ([]byte, []int) {
	return file_panacea_datadeal_v0_deal_proto_rawDescGZIP(), []int{2}
}

func (x *Certificate) GetUnsignedCertificate() *UnsignedCertificate {
	if x != nil {
		return x.UnsignedCertificate
	}
	return nil
}

func (x *Certificate) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

// UnsignedCertificate defines a certificate issued by an oracle as a result of data validation.
type UnsignedCertificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cid             string `protobuf:"bytes,1,opt,name=cid,proto3" json:"cid,omitempty"`
	UniqueId        string `protobuf:"bytes,2,opt,name=unique_id,proto3" json:"unique_id,omitempty"`
	OracleAddress   string `protobuf:"bytes,3,opt,name=oracle_address,proto3" json:"oracle_address,omitempty"`
	DealId          uint64 `protobuf:"varint,4,opt,name=deal_id,proto3" json:"deal_id,omitempty"`
	ProviderAddress string `protobuf:"bytes,5,opt,name=provider_address,proto3" json:"provider_address,omitempty"`
	DataHash        []byte `protobuf:"bytes,6,opt,name=data_hash,proto3" json:"data_hash,omitempty"`
}

func (x *UnsignedCertificate) Reset() {
	*x = UnsignedCertificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnsignedCertificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnsignedCertificate) ProtoMessage() {}

func (x *UnsignedCertificate) ProtoReflect() protoreflect.Message {
	mi := &file_panacea_datadeal_v0_deal_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnsignedCertificate.ProtoReflect.Descriptor instead.
func (*UnsignedCertificate) Descriptor() ([]byte, []int) {
	return file_panacea_datadeal_v0_deal_proto_rawDescGZIP(), []int{3}
}

func (x *UnsignedCertificate) GetCid() string {
	if x != nil {
		return x.Cid
	}
	return ""
}

func (x *UnsignedCertificate) GetUniqueId() string {
	if x != nil {
		return x.UniqueId
	}
	return ""
}

func (x *UnsignedCertificate) GetOracleAddress() string {
	if x != nil {
		return x.OracleAddress
	}
	return ""
}

func (x *UnsignedCertificate) GetDealId() uint64 {
	if x != nil {
		return x.DealId
	}
	return 0
}

func (x *UnsignedCertificate) GetProviderAddress() string {
	if x != nil {
		return x.ProviderAddress
	}
	return ""
}

func (x *UnsignedCertificate) GetDataHash() []byte {
	if x != nil {
		return x.DataHash
	}
	return nil
}

var File_panacea_datadeal_v0_deal_proto protoreflect.FileDescriptor

var file_panacea_datadeal_v0_deal_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x70, 0x61, 0x6e, 0x61, 0x63, 0x65, 0x61, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x64, 0x65,
	0x61, 0x6c, 0x2f, 0x76, 0x30, 0x2f, 0x64, 0x65, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x13, 0x70, 0x61, 0x6e, 0x61, 0x63, 0x65, 0x61, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x64, 0x65,
	0x61, 0x6c, 0x2e, 0x76, 0x30, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xa1, 0x01, 0x0a, 0x13, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x64,
	0x65, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x64, 0x65,
	0x61, 0x6c, 0x5f, 0x69, 0x64, 0x12, 0x2a, 0x0a, 0x10, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65,
	0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x10, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x12, 0x26, 0x0a, 0x0e, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x5f, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x65, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x65, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x61, 0x74,
	0x61, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x64, 0x61,
	0x74, 0x61, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x22, 0x5a, 0x0a, 0x14, 0x56, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x42, 0x0a, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70, 0x61, 0x6e, 0x61, 0x63, 0x65, 0x61, 0x2e, 0x64,
	0x61, 0x74, 0x61, 0x64, 0x65, 0x61, 0x6c, 0x2e, 0x76, 0x30, 0x2e, 0x43, 0x65, 0x72, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x22, 0x89, 0x01, 0x0a, 0x0b, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x12, 0x5c, 0x0a, 0x14, 0x75, 0x6e, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x5f,
	0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x28, 0x2e, 0x70, 0x61, 0x6e, 0x61, 0x63, 0x65, 0x61, 0x2e, 0x64, 0x61, 0x74, 0x61,
	0x64, 0x65, 0x61, 0x6c, 0x2e, 0x76, 0x30, 0x2e, 0x55, 0x6e, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64,
	0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x14, 0x75, 0x6e, 0x73,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x22,
	0xd1, 0x01, 0x0a, 0x13, 0x55, 0x6e, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x43, 0x65, 0x72, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x63, 0x69, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x75, 0x6e, 0x69,
	0x71, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x6e,
	0x69, 0x71, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x12, 0x26, 0x0a, 0x0e, 0x6f, 0x72, 0x61, 0x63, 0x6c,
	0x65, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x6f, 0x72, 0x61, 0x63, 0x6c, 0x65, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x64, 0x65, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x07, 0x64, 0x65, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x12, 0x2a, 0x0a, 0x10, 0x70, 0x72, 0x6f,
	0x76, 0x69, 0x64, 0x65, 0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x10, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x68, 0x61,
	0x73, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x68,
	0x61, 0x73, 0x68, 0x32, 0xa6, 0x01, 0x0a, 0x0f, 0x44, 0x61, 0x74, 0x61, 0x44, 0x65, 0x61, 0x6c,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x92, 0x01, 0x0a, 0x0c, 0x56, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x28, 0x2e, 0x70, 0x61, 0x6e, 0x61, 0x63,
	0x65, 0x61, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x64, 0x65, 0x61, 0x6c, 0x2e, 0x76, 0x30, 0x2e, 0x56,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x29, 0x2e, 0x70, 0x61, 0x6e, 0x61, 0x63, 0x65, 0x61, 0x2e, 0x64, 0x61, 0x74,
	0x61, 0x64, 0x65, 0x61, 0x6c, 0x2e, 0x76, 0x30, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2d, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x27, 0x22, 0x22, 0x2f, 0x76, 0x30, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2d,
	0x64, 0x65, 0x61, 0x6c, 0x2f, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2f, 0x7b, 0x64, 0x65, 0x61, 0x6c,
	0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x01, 0x2a, 0x42, 0x33, 0x5a, 0x31,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x64, 0x69, 0x62,
	0x6c, 0x6f, 0x63, 0x2f, 0x70, 0x61, 0x6e, 0x61, 0x63, 0x65, 0x61, 0x2d, 0x6f, 0x72, 0x61, 0x63,
	0x6c, 0x65, 0x2f, 0x70, 0x62, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x64, 0x65, 0x61, 0x6c, 0x2f, 0x76,
	0x30, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_panacea_datadeal_v0_deal_proto_rawDescOnce sync.Once
	file_panacea_datadeal_v0_deal_proto_rawDescData = file_panacea_datadeal_v0_deal_proto_rawDesc
)

func file_panacea_datadeal_v0_deal_proto_rawDescGZIP() []byte {
	file_panacea_datadeal_v0_deal_proto_rawDescOnce.Do(func() {
		file_panacea_datadeal_v0_deal_proto_rawDescData = protoimpl.X.CompressGZIP(file_panacea_datadeal_v0_deal_proto_rawDescData)
	})
	return file_panacea_datadeal_v0_deal_proto_rawDescData
}

var file_panacea_datadeal_v0_deal_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_panacea_datadeal_v0_deal_proto_goTypes = []interface{}{
	(*ValidateDataRequest)(nil),  // 0: panacea.datadeal.v0.ValidateDataRequest
	(*ValidateDataResponse)(nil), // 1: panacea.datadeal.v0.ValidateDataResponse
	(*Certificate)(nil),          // 2: panacea.datadeal.v0.Certificate
	(*UnsignedCertificate)(nil),  // 3: panacea.datadeal.v0.UnsignedCertificate
}
var file_panacea_datadeal_v0_deal_proto_depIdxs = []int32{
	2, // 0: panacea.datadeal.v0.ValidateDataResponse.certificate:type_name -> panacea.datadeal.v0.Certificate
	3, // 1: panacea.datadeal.v0.Certificate.unsigned_certificate:type_name -> panacea.datadeal.v0.UnsignedCertificate
	0, // 2: panacea.datadeal.v0.DataDealService.ValidateData:input_type -> panacea.datadeal.v0.ValidateDataRequest
	1, // 3: panacea.datadeal.v0.DataDealService.ValidateData:output_type -> panacea.datadeal.v0.ValidateDataResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_panacea_datadeal_v0_deal_proto_init() }
func file_panacea_datadeal_v0_deal_proto_init() {
	if File_panacea_datadeal_v0_deal_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_panacea_datadeal_v0_deal_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateDataRequest); i {
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
		file_panacea_datadeal_v0_deal_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateDataResponse); i {
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
		file_panacea_datadeal_v0_deal_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Certificate); i {
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
		file_panacea_datadeal_v0_deal_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnsignedCertificate); i {
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
			RawDescriptor: file_panacea_datadeal_v0_deal_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_panacea_datadeal_v0_deal_proto_goTypes,
		DependencyIndexes: file_panacea_datadeal_v0_deal_proto_depIdxs,
		MessageInfos:      file_panacea_datadeal_v0_deal_proto_msgTypes,
	}.Build()
	File_panacea_datadeal_v0_deal_proto = out.File
	file_panacea_datadeal_v0_deal_proto_rawDesc = nil
	file_panacea_datadeal_v0_deal_proto_goTypes = nil
	file_panacea_datadeal_v0_deal_proto_depIdxs = nil
}
