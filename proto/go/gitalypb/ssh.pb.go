// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ssh.proto

package gitalypb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type SSHUploadPackRequest struct {
	// 'repository' must be present in the first message.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// A chunk of raw data to be copied to 'git upload-pack' standard input
	Stdin []byte `protobuf:"bytes,2,opt,name=stdin,proto3" json:"stdin,omitempty"`
	// Parameters to use with git -c (key=value pairs)
	GitConfigOptions []string `protobuf:"bytes,4,rep,name=git_config_options,json=gitConfigOptions,proto3" json:"git_config_options,omitempty"`
	// Git protocol version
	GitProtocol          string   `protobuf:"bytes,5,opt,name=git_protocol,json=gitProtocol,proto3" json:"git_protocol,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSHUploadPackRequest) Reset()         { *m = SSHUploadPackRequest{} }
func (m *SSHUploadPackRequest) String() string { return proto.CompactTextString(m) }
func (*SSHUploadPackRequest) ProtoMessage()    {}
func (*SSHUploadPackRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef0eae71e2e883eb, []int{0}
}

func (m *SSHUploadPackRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadPackRequest.Unmarshal(m, b)
}
func (m *SSHUploadPackRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadPackRequest.Marshal(b, m, deterministic)
}
func (m *SSHUploadPackRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadPackRequest.Merge(m, src)
}
func (m *SSHUploadPackRequest) XXX_Size() int {
	return xxx_messageInfo_SSHUploadPackRequest.Size(m)
}
func (m *SSHUploadPackRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadPackRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadPackRequest proto.InternalMessageInfo

func (m *SSHUploadPackRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *SSHUploadPackRequest) GetStdin() []byte {
	if m != nil {
		return m.Stdin
	}
	return nil
}

func (m *SSHUploadPackRequest) GetGitConfigOptions() []string {
	if m != nil {
		return m.GitConfigOptions
	}
	return nil
}

func (m *SSHUploadPackRequest) GetGitProtocol() string {
	if m != nil {
		return m.GitProtocol
	}
	return ""
}

type SSHUploadPackResponse struct {
	// A chunk of raw data from 'git upload-pack' standard output
	Stdout []byte `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// A chunk of raw data from 'git upload-pack' standard error
	Stderr []byte `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// This field may be nil. This is intentional: only when the remote
	// command has finished can we return its exit status.
	ExitStatus           *ExitStatus `protobuf:"bytes,3,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SSHUploadPackResponse) Reset()         { *m = SSHUploadPackResponse{} }
func (m *SSHUploadPackResponse) String() string { return proto.CompactTextString(m) }
func (*SSHUploadPackResponse) ProtoMessage()    {}
func (*SSHUploadPackResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef0eae71e2e883eb, []int{1}
}

func (m *SSHUploadPackResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadPackResponse.Unmarshal(m, b)
}
func (m *SSHUploadPackResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadPackResponse.Marshal(b, m, deterministic)
}
func (m *SSHUploadPackResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadPackResponse.Merge(m, src)
}
func (m *SSHUploadPackResponse) XXX_Size() int {
	return xxx_messageInfo_SSHUploadPackResponse.Size(m)
}
func (m *SSHUploadPackResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadPackResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadPackResponse proto.InternalMessageInfo

func (m *SSHUploadPackResponse) GetStdout() []byte {
	if m != nil {
		return m.Stdout
	}
	return nil
}

func (m *SSHUploadPackResponse) GetStderr() []byte {
	if m != nil {
		return m.Stderr
	}
	return nil
}

func (m *SSHUploadPackResponse) GetExitStatus() *ExitStatus {
	if m != nil {
		return m.ExitStatus
	}
	return nil
}

type SSHReceivePackRequest struct {
	// 'repository' must be present in the first message.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// A chunk of raw data to be copied to 'git upload-pack' standard input
	Stdin []byte `protobuf:"bytes,2,opt,name=stdin,proto3" json:"stdin,omitempty"`
	// Contents of GL_ID, GL_REPOSITORY, and GL_USERNAME environment variables
	// for 'git receive-pack'
	GlId         string `protobuf:"bytes,3,opt,name=gl_id,json=glId,proto3" json:"gl_id,omitempty"`
	GlRepository string `protobuf:"bytes,4,opt,name=gl_repository,json=glRepository,proto3" json:"gl_repository,omitempty"`
	GlUsername   string `protobuf:"bytes,5,opt,name=gl_username,json=glUsername,proto3" json:"gl_username,omitempty"`
	// Git protocol version
	GitProtocol string `protobuf:"bytes,6,opt,name=git_protocol,json=gitProtocol,proto3" json:"git_protocol,omitempty"`
	// Parameters to use with git -c (key=value pairs)
	GitConfigOptions     []string `protobuf:"bytes,7,rep,name=git_config_options,json=gitConfigOptions,proto3" json:"git_config_options,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSHReceivePackRequest) Reset()         { *m = SSHReceivePackRequest{} }
func (m *SSHReceivePackRequest) String() string { return proto.CompactTextString(m) }
func (*SSHReceivePackRequest) ProtoMessage()    {}
func (*SSHReceivePackRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef0eae71e2e883eb, []int{2}
}

func (m *SSHReceivePackRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHReceivePackRequest.Unmarshal(m, b)
}
func (m *SSHReceivePackRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHReceivePackRequest.Marshal(b, m, deterministic)
}
func (m *SSHReceivePackRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHReceivePackRequest.Merge(m, src)
}
func (m *SSHReceivePackRequest) XXX_Size() int {
	return xxx_messageInfo_SSHReceivePackRequest.Size(m)
}
func (m *SSHReceivePackRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHReceivePackRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SSHReceivePackRequest proto.InternalMessageInfo

func (m *SSHReceivePackRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *SSHReceivePackRequest) GetStdin() []byte {
	if m != nil {
		return m.Stdin
	}
	return nil
}

func (m *SSHReceivePackRequest) GetGlId() string {
	if m != nil {
		return m.GlId
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGlRepository() string {
	if m != nil {
		return m.GlRepository
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGlUsername() string {
	if m != nil {
		return m.GlUsername
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGitProtocol() string {
	if m != nil {
		return m.GitProtocol
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGitConfigOptions() []string {
	if m != nil {
		return m.GitConfigOptions
	}
	return nil
}

type SSHReceivePackResponse struct {
	// A chunk of raw data from 'git receive-pack' standard output
	Stdout []byte `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// A chunk of raw data from 'git receive-pack' standard error
	Stderr []byte `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// This field may be nil. This is intentional: only when the remote
	// command has finished can we return its exit status.
	ExitStatus           *ExitStatus `protobuf:"bytes,3,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SSHReceivePackResponse) Reset()         { *m = SSHReceivePackResponse{} }
func (m *SSHReceivePackResponse) String() string { return proto.CompactTextString(m) }
func (*SSHReceivePackResponse) ProtoMessage()    {}
func (*SSHReceivePackResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef0eae71e2e883eb, []int{3}
}

func (m *SSHReceivePackResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHReceivePackResponse.Unmarshal(m, b)
}
func (m *SSHReceivePackResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHReceivePackResponse.Marshal(b, m, deterministic)
}
func (m *SSHReceivePackResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHReceivePackResponse.Merge(m, src)
}
func (m *SSHReceivePackResponse) XXX_Size() int {
	return xxx_messageInfo_SSHReceivePackResponse.Size(m)
}
func (m *SSHReceivePackResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHReceivePackResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SSHReceivePackResponse proto.InternalMessageInfo

func (m *SSHReceivePackResponse) GetStdout() []byte {
	if m != nil {
		return m.Stdout
	}
	return nil
}

func (m *SSHReceivePackResponse) GetStderr() []byte {
	if m != nil {
		return m.Stderr
	}
	return nil
}

func (m *SSHReceivePackResponse) GetExitStatus() *ExitStatus {
	if m != nil {
		return m.ExitStatus
	}
	return nil
}

type SSHUploadArchiveRequest struct {
	// 'repository' must be present in the first message.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// A chunk of raw data to be copied to 'git upload-archive' standard input
	Stdin                []byte   `protobuf:"bytes,2,opt,name=stdin,proto3" json:"stdin,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSHUploadArchiveRequest) Reset()         { *m = SSHUploadArchiveRequest{} }
func (m *SSHUploadArchiveRequest) String() string { return proto.CompactTextString(m) }
func (*SSHUploadArchiveRequest) ProtoMessage()    {}
func (*SSHUploadArchiveRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef0eae71e2e883eb, []int{4}
}

func (m *SSHUploadArchiveRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadArchiveRequest.Unmarshal(m, b)
}
func (m *SSHUploadArchiveRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadArchiveRequest.Marshal(b, m, deterministic)
}
func (m *SSHUploadArchiveRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadArchiveRequest.Merge(m, src)
}
func (m *SSHUploadArchiveRequest) XXX_Size() int {
	return xxx_messageInfo_SSHUploadArchiveRequest.Size(m)
}
func (m *SSHUploadArchiveRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadArchiveRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadArchiveRequest proto.InternalMessageInfo

func (m *SSHUploadArchiveRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *SSHUploadArchiveRequest) GetStdin() []byte {
	if m != nil {
		return m.Stdin
	}
	return nil
}

type SSHUploadArchiveResponse struct {
	// A chunk of raw data from 'git upload-archive' standard output
	Stdout []byte `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// A chunk of raw data from 'git upload-archive' standard error
	Stderr []byte `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// This value will only be set on the last message
	ExitStatus           *ExitStatus `protobuf:"bytes,3,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SSHUploadArchiveResponse) Reset()         { *m = SSHUploadArchiveResponse{} }
func (m *SSHUploadArchiveResponse) String() string { return proto.CompactTextString(m) }
func (*SSHUploadArchiveResponse) ProtoMessage()    {}
func (*SSHUploadArchiveResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef0eae71e2e883eb, []int{5}
}

func (m *SSHUploadArchiveResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadArchiveResponse.Unmarshal(m, b)
}
func (m *SSHUploadArchiveResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadArchiveResponse.Marshal(b, m, deterministic)
}
func (m *SSHUploadArchiveResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadArchiveResponse.Merge(m, src)
}
func (m *SSHUploadArchiveResponse) XXX_Size() int {
	return xxx_messageInfo_SSHUploadArchiveResponse.Size(m)
}
func (m *SSHUploadArchiveResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadArchiveResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadArchiveResponse proto.InternalMessageInfo

func (m *SSHUploadArchiveResponse) GetStdout() []byte {
	if m != nil {
		return m.Stdout
	}
	return nil
}

func (m *SSHUploadArchiveResponse) GetStderr() []byte {
	if m != nil {
		return m.Stderr
	}
	return nil
}

func (m *SSHUploadArchiveResponse) GetExitStatus() *ExitStatus {
	if m != nil {
		return m.ExitStatus
	}
	return nil
}

func init() {
	proto.RegisterType((*SSHUploadPackRequest)(nil), "gitaly.SSHUploadPackRequest")
	proto.RegisterType((*SSHUploadPackResponse)(nil), "gitaly.SSHUploadPackResponse")
	proto.RegisterType((*SSHReceivePackRequest)(nil), "gitaly.SSHReceivePackRequest")
	proto.RegisterType((*SSHReceivePackResponse)(nil), "gitaly.SSHReceivePackResponse")
	proto.RegisterType((*SSHUploadArchiveRequest)(nil), "gitaly.SSHUploadArchiveRequest")
	proto.RegisterType((*SSHUploadArchiveResponse)(nil), "gitaly.SSHUploadArchiveResponse")
}

func init() { proto.RegisterFile("ssh.proto", fileDescriptor_ef0eae71e2e883eb) }

var fileDescriptor_ef0eae71e2e883eb = []byte{
	// 506 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x53, 0xcf, 0x6e, 0xd3, 0x30,
	0x18, 0x57, 0xda, 0xb4, 0xac, 0x5f, 0x3b, 0x54, 0x99, 0x6d, 0x44, 0x11, 0xb0, 0x10, 0x2e, 0x39,
	0x8c, 0x76, 0xda, 0x2e, 0x5c, 0x19, 0x42, 0x1a, 0x5c, 0x98, 0x1c, 0x4d, 0x42, 0x70, 0x88, 0xdc,
	0xc4, 0xb8, 0x16, 0x6e, 0x1c, 0x6c, 0xb7, 0xda, 0x24, 0x10, 0x0f, 0xc1, 0x01, 0x9e, 0x80, 0x47,
	0xe1, 0x09, 0x78, 0x1a, 0x4e, 0x68, 0x4e, 0x28, 0x4d, 0xb3, 0x1e, 0xb7, 0x9b, 0xbf, 0xdf, 0xef,
	0xf3, 0xf7, 0xe7, 0xf7, 0xb3, 0xa1, 0xa7, 0xf5, 0x74, 0x54, 0x28, 0x69, 0x24, 0xea, 0x32, 0x6e,
	0x88, 0xb8, 0xf4, 0x41, 0xf0, 0xdc, 0x94, 0x98, 0x3f, 0xd0, 0x53, 0xa2, 0x68, 0x56, 0x46, 0xe1,
	0x6f, 0x07, 0x76, 0xe2, 0xf8, 0xf4, 0xbc, 0x10, 0x92, 0x64, 0x67, 0x24, 0xfd, 0x88, 0xe9, 0xa7,
	0x39, 0xd5, 0x06, 0x3d, 0x03, 0x50, 0xb4, 0x90, 0x9a, 0x1b, 0xa9, 0x2e, 0x3d, 0x27, 0x70, 0xa2,
	0xfe, 0x11, 0x1a, 0x95, 0xf5, 0x46, 0x78, 0xc9, 0x9c, 0xb8, 0x3f, 0x7e, 0x1d, 0x38, 0x78, 0x25,
	0x17, 0xed, 0x40, 0x47, 0x9b, 0x8c, 0xe7, 0x5e, 0x2b, 0x70, 0xa2, 0x01, 0x2e, 0x03, 0x74, 0x00,
	0x88, 0x71, 0x93, 0xa4, 0x32, 0xff, 0xc0, 0x59, 0x22, 0x0b, 0xc3, 0x65, 0xae, 0x3d, 0x37, 0x68,
	0x47, 0x3d, 0x3c, 0x64, 0xdc, 0xbc, 0xb0, 0xc4, 0x9b, 0x12, 0x47, 0x8f, 0x61, 0x70, 0x95, 0x6d,
	0x67, 0x4c, 0xa5, 0xf0, 0x3a, 0x81, 0x13, 0xf5, 0x70, 0x9f, 0x71, 0x73, 0x56, 0x41, 0xaf, 0xdd,
	0xad, 0xf6, 0xd0, 0xc5, 0xbb, 0x2b, 0x45, 0x0b, 0xa2, 0xc8, 0x8c, 0x1a, 0xaa, 0x74, 0xf8, 0x19,
	0x76, 0xd7, 0xb6, 0xd2, 0x85, 0xcc, 0x35, 0x45, 0x7b, 0xd0, 0xd5, 0x26, 0x93, 0x73, 0x63, 0x57,
	0x1a, 0xe0, 0x2a, 0xaa, 0x70, 0xaa, 0x54, 0x35, 0x75, 0x15, 0xa1, 0x63, 0xe8, 0xd3, 0x0b, 0x6e,
	0x12, 0x6d, 0x88, 0x99, 0x6b, 0xaf, 0x5d, 0xd7, 0xe1, 0xe5, 0x05, 0x37, 0xb1, 0x65, 0x30, 0xd0,
	0xe5, 0x39, 0xfc, 0xd6, 0xb2, 0xed, 0x31, 0x4d, 0x29, 0x5f, 0xd0, 0x9b, 0x54, 0xf5, 0x1e, 0x74,
	0x98, 0x48, 0x78, 0x66, 0x07, 0xeb, 0x61, 0x97, 0x89, 0x57, 0x19, 0x7a, 0x02, 0xdb, 0x4c, 0x24,
	0x2b, 0x7d, 0x5c, 0x4b, 0x0e, 0x98, 0xf8, 0xdf, 0x01, 0xed, 0x43, 0x9f, 0x89, 0x64, 0xae, 0xa9,
	0xca, 0xc9, 0x8c, 0x56, 0x02, 0x03, 0x13, 0xe7, 0x15, 0xd2, 0xb0, 0xa0, 0xdb, 0xb0, 0x60, 0x83,
	0xa7, 0x77, 0xae, 0xf7, 0x34, 0xfc, 0x02, 0x7b, 0xeb, 0xa2, 0xdc, 0xa6, 0x29, 0x1c, 0xee, 0x2f,
	0x9f, 0xc4, 0x73, 0x95, 0x4e, 0xf9, 0x82, 0xde, 0x90, 0x2b, 0xe1, 0x57, 0xf0, 0x9a, 0xad, 0x6e,
	0x71, 0xd7, 0xa3, 0x9f, 0x2d, 0x80, 0x38, 0x3e, 0x8d, 0xa9, 0x5a, 0xf0, 0x94, 0xa2, 0xb7, 0xb0,
	0x5d, 0xfb, 0x0d, 0xe8, 0xc1, 0xbf, 0xfb, 0xd7, 0x7d, 0x7d, 0xff, 0xe1, 0x06, 0xb6, 0xdc, 0x20,
	0xec, 0xfe, 0xf9, 0x1e, 0xb5, 0xb6, 0x5a, 0x91, 0x73, 0xe8, 0xa0, 0xf7, 0x70, 0xb7, 0xee, 0x29,
	0x5a, 0xbd, 0xdc, 0xfc, 0x00, 0xfe, 0xa3, 0x4d, 0x74, 0xad, 0xb8, 0x63, 0x8b, 0x13, 0x18, 0xae,
	0xcb, 0x88, 0xf6, 0x1b, 0xb3, 0xd5, 0xbd, 0xf4, 0x83, 0xcd, 0x09, 0xcd, 0xf9, 0x4f, 0x0e, 0xdf,
	0x5d, 0xa5, 0x0b, 0x32, 0x19, 0xa5, 0x72, 0x36, 0x2e, 0x8f, 0x4f, 0xa5, 0x62, 0xe3, 0xb2, 0xc8,
	0xd8, 0xbe, 0xfe, 0x31, 0x93, 0x55, 0x5c, 0x4c, 0x26, 0x5d, 0x0b, 0x1d, 0xff, 0x0d, 0x00, 0x00,
	0xff, 0xff, 0x38, 0x5c, 0x3c, 0x09, 0x66, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SSHServiceClient is the client API for SSHService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SSHServiceClient interface {
	// To forward 'git upload-pack' to Gitaly for SSH sessions
	SSHUploadPack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadPackClient, error)
	// To forward 'git receive-pack' to Gitaly for SSH sessions
	SSHReceivePack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHReceivePackClient, error)
	// To forward 'git upload-archive' to Gitaly for SSH sessions
	SSHUploadArchive(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadArchiveClient, error)
}

type sSHServiceClient struct {
	cc *grpc.ClientConn
}

func NewSSHServiceClient(cc *grpc.ClientConn) SSHServiceClient {
	return &sSHServiceClient{cc}
}

func (c *sSHServiceClient) SSHUploadPack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadPackClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SSHService_serviceDesc.Streams[0], "/gitaly.SSHService/SSHUploadPack", opts...)
	if err != nil {
		return nil, err
	}
	x := &sSHServiceSSHUploadPackClient{stream}
	return x, nil
}

type SSHService_SSHUploadPackClient interface {
	Send(*SSHUploadPackRequest) error
	Recv() (*SSHUploadPackResponse, error)
	grpc.ClientStream
}

type sSHServiceSSHUploadPackClient struct {
	grpc.ClientStream
}

func (x *sSHServiceSSHUploadPackClient) Send(m *SSHUploadPackRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadPackClient) Recv() (*SSHUploadPackResponse, error) {
	m := new(SSHUploadPackResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sSHServiceClient) SSHReceivePack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHReceivePackClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SSHService_serviceDesc.Streams[1], "/gitaly.SSHService/SSHReceivePack", opts...)
	if err != nil {
		return nil, err
	}
	x := &sSHServiceSSHReceivePackClient{stream}
	return x, nil
}

type SSHService_SSHReceivePackClient interface {
	Send(*SSHReceivePackRequest) error
	Recv() (*SSHReceivePackResponse, error)
	grpc.ClientStream
}

type sSHServiceSSHReceivePackClient struct {
	grpc.ClientStream
}

func (x *sSHServiceSSHReceivePackClient) Send(m *SSHReceivePackRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *sSHServiceSSHReceivePackClient) Recv() (*SSHReceivePackResponse, error) {
	m := new(SSHReceivePackResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sSHServiceClient) SSHUploadArchive(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadArchiveClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SSHService_serviceDesc.Streams[2], "/gitaly.SSHService/SSHUploadArchive", opts...)
	if err != nil {
		return nil, err
	}
	x := &sSHServiceSSHUploadArchiveClient{stream}
	return x, nil
}

type SSHService_SSHUploadArchiveClient interface {
	Send(*SSHUploadArchiveRequest) error
	Recv() (*SSHUploadArchiveResponse, error)
	grpc.ClientStream
}

type sSHServiceSSHUploadArchiveClient struct {
	grpc.ClientStream
}

func (x *sSHServiceSSHUploadArchiveClient) Send(m *SSHUploadArchiveRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadArchiveClient) Recv() (*SSHUploadArchiveResponse, error) {
	m := new(SSHUploadArchiveResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SSHServiceServer is the server API for SSHService service.
type SSHServiceServer interface {
	// To forward 'git upload-pack' to Gitaly for SSH sessions
	SSHUploadPack(SSHService_SSHUploadPackServer) error
	// To forward 'git receive-pack' to Gitaly for SSH sessions
	SSHReceivePack(SSHService_SSHReceivePackServer) error
	// To forward 'git upload-archive' to Gitaly for SSH sessions
	SSHUploadArchive(SSHService_SSHUploadArchiveServer) error
}

// UnimplementedSSHServiceServer can be embedded to have forward compatible implementations.
type UnimplementedSSHServiceServer struct {
}

func (*UnimplementedSSHServiceServer) SSHUploadPack(srv SSHService_SSHUploadPackServer) error {
	return status.Errorf(codes.Unimplemented, "method SSHUploadPack not implemented")
}
func (*UnimplementedSSHServiceServer) SSHReceivePack(srv SSHService_SSHReceivePackServer) error {
	return status.Errorf(codes.Unimplemented, "method SSHReceivePack not implemented")
}
func (*UnimplementedSSHServiceServer) SSHUploadArchive(srv SSHService_SSHUploadArchiveServer) error {
	return status.Errorf(codes.Unimplemented, "method SSHUploadArchive not implemented")
}

func RegisterSSHServiceServer(s *grpc.Server, srv SSHServiceServer) {
	s.RegisterService(&_SSHService_serviceDesc, srv)
}

func _SSHService_SSHUploadPack_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SSHServiceServer).SSHUploadPack(&sSHServiceSSHUploadPackServer{stream})
}

type SSHService_SSHUploadPackServer interface {
	Send(*SSHUploadPackResponse) error
	Recv() (*SSHUploadPackRequest, error)
	grpc.ServerStream
}

type sSHServiceSSHUploadPackServer struct {
	grpc.ServerStream
}

func (x *sSHServiceSSHUploadPackServer) Send(m *SSHUploadPackResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadPackServer) Recv() (*SSHUploadPackRequest, error) {
	m := new(SSHUploadPackRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SSHService_SSHReceivePack_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SSHServiceServer).SSHReceivePack(&sSHServiceSSHReceivePackServer{stream})
}

type SSHService_SSHReceivePackServer interface {
	Send(*SSHReceivePackResponse) error
	Recv() (*SSHReceivePackRequest, error)
	grpc.ServerStream
}

type sSHServiceSSHReceivePackServer struct {
	grpc.ServerStream
}

func (x *sSHServiceSSHReceivePackServer) Send(m *SSHReceivePackResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *sSHServiceSSHReceivePackServer) Recv() (*SSHReceivePackRequest, error) {
	m := new(SSHReceivePackRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SSHService_SSHUploadArchive_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SSHServiceServer).SSHUploadArchive(&sSHServiceSSHUploadArchiveServer{stream})
}

type SSHService_SSHUploadArchiveServer interface {
	Send(*SSHUploadArchiveResponse) error
	Recv() (*SSHUploadArchiveRequest, error)
	grpc.ServerStream
}

type sSHServiceSSHUploadArchiveServer struct {
	grpc.ServerStream
}

func (x *sSHServiceSSHUploadArchiveServer) Send(m *SSHUploadArchiveResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadArchiveServer) Recv() (*SSHUploadArchiveRequest, error) {
	m := new(SSHUploadArchiveRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _SSHService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.SSHService",
	HandlerType: (*SSHServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SSHUploadPack",
			Handler:       _SSHService_SSHUploadPack_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "SSHReceivePack",
			Handler:       _SSHService_SSHReceivePack_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "SSHUploadArchive",
			Handler:       _SSHService_SSHUploadArchive_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "ssh.proto",
}
