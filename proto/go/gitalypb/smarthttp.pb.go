// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.1
// source: smarthttp.proto

package gitalypb

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

// InfoRefsRequest is a request for the InfoRefsUploadPack and InfoRefsUploadPack rpcs.
type InfoRefsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Repository is the repository on which to operate.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// GitConfigOptions are parameters to use with git -c (key=value pairs).
	GitConfigOptions []string `protobuf:"bytes,2,rep,name=git_config_options,json=gitConfigOptions,proto3" json:"git_config_options,omitempty"`
	// GitProtocol is the git protocol version.
	GitProtocol string `protobuf:"bytes,3,opt,name=git_protocol,json=gitProtocol,proto3" json:"git_protocol,omitempty"`
}

func (x *InfoRefsRequest) Reset() {
	*x = InfoRefsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarthttp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfoRefsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoRefsRequest) ProtoMessage() {}

func (x *InfoRefsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_smarthttp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoRefsRequest.ProtoReflect.Descriptor instead.
func (*InfoRefsRequest) Descriptor() ([]byte, []int) {
	return file_smarthttp_proto_rawDescGZIP(), []int{0}
}

func (x *InfoRefsRequest) GetRepository() *Repository {
	if x != nil {
		return x.Repository
	}
	return nil
}

func (x *InfoRefsRequest) GetGitConfigOptions() []string {
	if x != nil {
		return x.GitConfigOptions
	}
	return nil
}

func (x *InfoRefsRequest) GetGitProtocol() string {
	if x != nil {
		return x.GitProtocol
	}
	return ""
}

// InfoRefsResponse is the response of InfoRefsUploadPack and InfoRefsUploadPack rpcs.
// It is used to provide the client with the servers advertised refs.
type InfoRefsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Data is the raw data copied from the stdout of git-upload-pack(1) or
	// git-receive-pack(1) when used with the `--advertise-refs` flag.
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *InfoRefsResponse) Reset() {
	*x = InfoRefsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarthttp_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfoRefsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoRefsResponse) ProtoMessage() {}

func (x *InfoRefsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_smarthttp_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoRefsResponse.ProtoReflect.Descriptor instead.
func (*InfoRefsResponse) Descriptor() ([]byte, []int) {
	return file_smarthttp_proto_rawDescGZIP(), []int{1}
}

func (x *InfoRefsResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// PostUploadPackWithSidechannelRequest is the request for the PostUploadPackWithSidechannel rpc.
type PostUploadPackWithSidechannelRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Repository is the repository on which to operate.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// GitConfigOptions are parameters to use with git -c (key=value pairs).
	GitConfigOptions []string `protobuf:"bytes,2,rep,name=git_config_options,json=gitConfigOptions,proto3" json:"git_config_options,omitempty"`
	// GitProtocol is the git protocol version.
	GitProtocol string `protobuf:"bytes,3,opt,name=git_protocol,json=gitProtocol,proto3" json:"git_protocol,omitempty"`
}

func (x *PostUploadPackWithSidechannelRequest) Reset() {
	*x = PostUploadPackWithSidechannelRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarthttp_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostUploadPackWithSidechannelRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostUploadPackWithSidechannelRequest) ProtoMessage() {}

func (x *PostUploadPackWithSidechannelRequest) ProtoReflect() protoreflect.Message {
	mi := &file_smarthttp_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostUploadPackWithSidechannelRequest.ProtoReflect.Descriptor instead.
func (*PostUploadPackWithSidechannelRequest) Descriptor() ([]byte, []int) {
	return file_smarthttp_proto_rawDescGZIP(), []int{2}
}

func (x *PostUploadPackWithSidechannelRequest) GetRepository() *Repository {
	if x != nil {
		return x.Repository
	}
	return nil
}

func (x *PostUploadPackWithSidechannelRequest) GetGitConfigOptions() []string {
	if x != nil {
		return x.GitConfigOptions
	}
	return nil
}

func (x *PostUploadPackWithSidechannelRequest) GetGitProtocol() string {
	if x != nil {
		return x.GitProtocol
	}
	return ""
}

// PostUploadPackWithSidechannelResponse is the response for the PostUploadPackWithSidechannel rpc.
// This is an empty response since the raw data is transferred to the client via the sidechannel
// exclusively.
type PostUploadPackWithSidechannelResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Packfile negotiation statistics.
	PackfileNegotiationStatistics *PackfileNegotiationStatistics `protobuf:"bytes,1,opt,name=packfile_negotiation_statistics,json=packfileNegotiationStatistics,proto3" json:"packfile_negotiation_statistics,omitempty"`
}

func (x *PostUploadPackWithSidechannelResponse) Reset() {
	*x = PostUploadPackWithSidechannelResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarthttp_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostUploadPackWithSidechannelResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostUploadPackWithSidechannelResponse) ProtoMessage() {}

func (x *PostUploadPackWithSidechannelResponse) ProtoReflect() protoreflect.Message {
	mi := &file_smarthttp_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostUploadPackWithSidechannelResponse.ProtoReflect.Descriptor instead.
func (*PostUploadPackWithSidechannelResponse) Descriptor() ([]byte, []int) {
	return file_smarthttp_proto_rawDescGZIP(), []int{3}
}

func (x *PostUploadPackWithSidechannelResponse) GetPackfileNegotiationStatistics() *PackfileNegotiationStatistics {
	if x != nil {
		return x.PackfileNegotiationStatistics
	}
	return nil
}

// PostReceivePackRequest is the request for the PostReceivePack rpc. It is a stream used to
// transfer the raw data from the client to the servers stdin of git-receive-pack(1) process.
type PostReceivePackRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Repository is the repository on which to operate.
	// It should only be present in the first message of the stream.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// Data is the raw data to be copied to stdin of 'git receive-pack'.
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	// GlID is the GitLab ID of the user. This is used by Git {pre,post}-receive hooks.
	// It should only be present in the first message of the stream.
	GlId string `protobuf:"bytes,3,opt,name=gl_id,json=glId,proto3" json:"gl_id,omitempty"`
	// GlRepository refers to the GitLab repository. This is used by Git {pre,post}-receive hooks.
	// It should only be present in the first message of the stream.
	GlRepository string `protobuf:"bytes,4,opt,name=gl_repository,json=glRepository,proto3" json:"gl_repository,omitempty"`
	// GlID is the GitLab Username of the user. This is used by Git {pre,post}-receive hooks.
	// It should only be present in the first message of the stream.
	GlUsername string `protobuf:"bytes,5,opt,name=gl_username,json=glUsername,proto3" json:"gl_username,omitempty"`
	// GitProtocol is the git protocol version.
	GitProtocol string `protobuf:"bytes,6,opt,name=git_protocol,json=gitProtocol,proto3" json:"git_protocol,omitempty"`
	// GitConfigOptions are parameters to use with git -c (key=value pairs).
	GitConfigOptions []string `protobuf:"bytes,7,rep,name=git_config_options,json=gitConfigOptions,proto3" json:"git_config_options,omitempty"`
}

func (x *PostReceivePackRequest) Reset() {
	*x = PostReceivePackRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarthttp_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostReceivePackRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostReceivePackRequest) ProtoMessage() {}

func (x *PostReceivePackRequest) ProtoReflect() protoreflect.Message {
	mi := &file_smarthttp_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostReceivePackRequest.ProtoReflect.Descriptor instead.
func (*PostReceivePackRequest) Descriptor() ([]byte, []int) {
	return file_smarthttp_proto_rawDescGZIP(), []int{4}
}

func (x *PostReceivePackRequest) GetRepository() *Repository {
	if x != nil {
		return x.Repository
	}
	return nil
}

func (x *PostReceivePackRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *PostReceivePackRequest) GetGlId() string {
	if x != nil {
		return x.GlId
	}
	return ""
}

func (x *PostReceivePackRequest) GetGlRepository() string {
	if x != nil {
		return x.GlRepository
	}
	return ""
}

func (x *PostReceivePackRequest) GetGlUsername() string {
	if x != nil {
		return x.GlUsername
	}
	return ""
}

func (x *PostReceivePackRequest) GetGitProtocol() string {
	if x != nil {
		return x.GitProtocol
	}
	return ""
}

func (x *PostReceivePackRequest) GetGitConfigOptions() []string {
	if x != nil {
		return x.GitConfigOptions
	}
	return nil
}

// PostReceivePackResponse is the response for the PostReceivePack rpc. It is a stream used to
// transfer the raw data from the stdout of git-receive-pack(1) from the server to the client.
type PostReceivePackResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Data is the raw data from the stdout of 'git receive-pack'.
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *PostReceivePackResponse) Reset() {
	*x = PostReceivePackResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarthttp_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostReceivePackResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostReceivePackResponse) ProtoMessage() {}

func (x *PostReceivePackResponse) ProtoReflect() protoreflect.Message {
	mi := &file_smarthttp_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostReceivePackResponse.ProtoReflect.Descriptor instead.
func (*PostReceivePackResponse) Descriptor() ([]byte, []int) {
	return file_smarthttp_proto_rawDescGZIP(), []int{5}
}

func (x *PostReceivePackResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_smarthttp_proto protoreflect.FileDescriptor

var file_smarthttp_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x68, 0x74, 0x74, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x1a, 0x0a, 0x6c, 0x69, 0x6e, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x70, 0x61, 0x63, 0x6b, 0x66, 0x69, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x9c, 0x01, 0x0a, 0x0f, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x66, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x38, 0x0a, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73,
	0x69, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x69,
	0x74, 0x61, 0x6c, 0x79, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x42,
	0x04, 0x98, 0xc6, 0x2c, 0x01, 0x52, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72,
	0x79, 0x12, 0x2c, 0x0a, 0x12, 0x67, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x10, 0x67,
	0x69, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x21, 0x0a, 0x0c, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x67, 0x69, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x22, 0x26, 0x0a, 0x10, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x66, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0xb1, 0x01, 0x0a, 0x24, 0x50,
	0x6f, 0x73, 0x74, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x57, 0x69, 0x74,
	0x68, 0x53, 0x69, 0x64, 0x65, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x38, 0x0a, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79,
	0x2e, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x04, 0x98, 0xc6, 0x2c,
	0x01, 0x52, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x2c, 0x0a,
	0x12, 0x67, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x10, 0x67, 0x69, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x67,
	0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x67, 0x69, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0x96,
	0x01, 0x0a, 0x25, 0x50, 0x6f, 0x73, 0x74, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x61, 0x63,
	0x6b, 0x57, 0x69, 0x74, 0x68, 0x53, 0x69, 0x64, 0x65, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6d, 0x0a, 0x1f, 0x70, 0x61, 0x63, 0x6b,
	0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6e, 0x65, 0x67, 0x6f, 0x74, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x66,
	0x69, 0x6c, 0x65, 0x4e, 0x65, 0x67, 0x6f, 0x74, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x52, 0x1d, 0x70, 0x61, 0x63, 0x6b, 0x66, 0x69,
	0x6c, 0x65, 0x4e, 0x65, 0x67, 0x6f, 0x74, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61,
	0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x22, 0x92, 0x02, 0x0a, 0x16, 0x50, 0x6f, 0x73, 0x74,
	0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x38, 0x0a, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e,
	0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x04, 0x98, 0xc6, 0x2c, 0x01,
	0x52, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x13, 0x0a, 0x05, 0x67, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x67, 0x6c, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x67, 0x6c, 0x5f, 0x72, 0x65, 0x70, 0x6f,
	0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x67, 0x6c,
	0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x1f, 0x0a, 0x0b, 0x67, 0x6c,
	0x5f, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x67, 0x6c, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x67,
	0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x67, 0x69, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x2c,
	0x0a, 0x12, 0x67, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x10, 0x67, 0x69, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2d, 0x0a, 0x17,
	0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32, 0xa0, 0x03, 0x0a, 0x10,
	0x53, 0x6d, 0x61, 0x72, 0x74, 0x48, 0x54, 0x54, 0x50, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x51, 0x0a, 0x12, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x66, 0x73, 0x55, 0x70, 0x6c, 0x6f,
	0x61, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x12, 0x17, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x66, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x18, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x66,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x06, 0xfa, 0x97, 0x28, 0x02, 0x08,
	0x02, 0x30, 0x01, 0x12, 0x52, 0x0a, 0x13, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x66, 0x73, 0x52,
	0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x12, 0x17, 0x2e, 0x67, 0x69, 0x74,
	0x61, 0x6c, 0x79, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x66, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x65, 0x66, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x06, 0xfa,
	0x97, 0x28, 0x02, 0x08, 0x02, 0x30, 0x01, 0x12, 0x84, 0x01, 0x0a, 0x1d, 0x50, 0x6f, 0x73, 0x74,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x57, 0x69, 0x74, 0x68, 0x53, 0x69,
	0x64, 0x65, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x2c, 0x2e, 0x67, 0x69, 0x74, 0x61,
	0x6c, 0x79, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x61, 0x63,
	0x6b, 0x57, 0x69, 0x74, 0x68, 0x53, 0x69, 0x64, 0x65, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79,
	0x2e, 0x50, 0x6f, 0x73, 0x74, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x57,
	0x69, 0x74, 0x68, 0x53, 0x69, 0x64, 0x65, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x06, 0xfa, 0x97, 0x28, 0x02, 0x08, 0x02, 0x12, 0x5e,
	0x0a, 0x0f, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x50, 0x61, 0x63,
	0x6b, 0x12, 0x1e, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x52,
	0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1f, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x52,
	0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x06, 0xfa, 0x97, 0x28, 0x02, 0x08, 0x01, 0x28, 0x01, 0x30, 0x01, 0x42, 0x34,
	0x5a, 0x32, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x69, 0x74,
	0x6c, 0x61, 0x62, 0x2d, 0x6f, 0x72, 0x67, 0x2f, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2f, 0x76,
	0x31, 0x36, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x67, 0x69, 0x74, 0x61,
	0x6c, 0x79, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_smarthttp_proto_rawDescOnce sync.Once
	file_smarthttp_proto_rawDescData = file_smarthttp_proto_rawDesc
)

func file_smarthttp_proto_rawDescGZIP() []byte {
	file_smarthttp_proto_rawDescOnce.Do(func() {
		file_smarthttp_proto_rawDescData = protoimpl.X.CompressGZIP(file_smarthttp_proto_rawDescData)
	})
	return file_smarthttp_proto_rawDescData
}

var file_smarthttp_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_smarthttp_proto_goTypes = []interface{}{
	(*InfoRefsRequest)(nil),                       // 0: gitaly.InfoRefsRequest
	(*InfoRefsResponse)(nil),                      // 1: gitaly.InfoRefsResponse
	(*PostUploadPackWithSidechannelRequest)(nil),  // 2: gitaly.PostUploadPackWithSidechannelRequest
	(*PostUploadPackWithSidechannelResponse)(nil), // 3: gitaly.PostUploadPackWithSidechannelResponse
	(*PostReceivePackRequest)(nil),                // 4: gitaly.PostReceivePackRequest
	(*PostReceivePackResponse)(nil),               // 5: gitaly.PostReceivePackResponse
	(*Repository)(nil),                            // 6: gitaly.Repository
	(*PackfileNegotiationStatistics)(nil),         // 7: gitaly.PackfileNegotiationStatistics
}
var file_smarthttp_proto_depIdxs = []int32{
	6, // 0: gitaly.InfoRefsRequest.repository:type_name -> gitaly.Repository
	6, // 1: gitaly.PostUploadPackWithSidechannelRequest.repository:type_name -> gitaly.Repository
	7, // 2: gitaly.PostUploadPackWithSidechannelResponse.packfile_negotiation_statistics:type_name -> gitaly.PackfileNegotiationStatistics
	6, // 3: gitaly.PostReceivePackRequest.repository:type_name -> gitaly.Repository
	0, // 4: gitaly.SmartHTTPService.InfoRefsUploadPack:input_type -> gitaly.InfoRefsRequest
	0, // 5: gitaly.SmartHTTPService.InfoRefsReceivePack:input_type -> gitaly.InfoRefsRequest
	2, // 6: gitaly.SmartHTTPService.PostUploadPackWithSidechannel:input_type -> gitaly.PostUploadPackWithSidechannelRequest
	4, // 7: gitaly.SmartHTTPService.PostReceivePack:input_type -> gitaly.PostReceivePackRequest
	1, // 8: gitaly.SmartHTTPService.InfoRefsUploadPack:output_type -> gitaly.InfoRefsResponse
	1, // 9: gitaly.SmartHTTPService.InfoRefsReceivePack:output_type -> gitaly.InfoRefsResponse
	3, // 10: gitaly.SmartHTTPService.PostUploadPackWithSidechannel:output_type -> gitaly.PostUploadPackWithSidechannelResponse
	5, // 11: gitaly.SmartHTTPService.PostReceivePack:output_type -> gitaly.PostReceivePackResponse
	8, // [8:12] is the sub-list for method output_type
	4, // [4:8] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_smarthttp_proto_init() }
func file_smarthttp_proto_init() {
	if File_smarthttp_proto != nil {
		return
	}
	file_lint_proto_init()
	file_packfile_proto_init()
	file_shared_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_smarthttp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InfoRefsRequest); i {
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
		file_smarthttp_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InfoRefsResponse); i {
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
		file_smarthttp_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PostUploadPackWithSidechannelRequest); i {
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
		file_smarthttp_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PostUploadPackWithSidechannelResponse); i {
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
		file_smarthttp_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PostReceivePackRequest); i {
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
		file_smarthttp_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PostReceivePackResponse); i {
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
			RawDescriptor: file_smarthttp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_smarthttp_proto_goTypes,
		DependencyIndexes: file_smarthttp_proto_depIdxs,
		MessageInfos:      file_smarthttp_proto_msgTypes,
	}.Build()
	File_smarthttp_proto = out.File
	file_smarthttp_proto_rawDesc = nil
	file_smarthttp_proto_goTypes = nil
	file_smarthttp_proto_depIdxs = nil
}
