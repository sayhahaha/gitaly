// Code generated by protoc-gen-go. DO NOT EDIT.
// source: blob.proto

/*
Package gitaly is a generated protocol buffer package.

It is generated from these files:
	blob.proto
	commit.proto
	conflicts.proto
	deprecated-services.proto
	diff.proto
	namespace.proto
	notifications.proto
	operations.proto
	ref.proto
	remote.proto
	repository-service.proto
	server.proto
	shared.proto
	smarthttp.proto
	ssh.proto
	wiki.proto

It has these top-level messages:
	GetBlobRequest
	GetBlobResponse
	GetBlobsRequest
	GetBlobsResponse
	LFSPointer
	GetLFSPointersRequest
	GetLFSPointersResponse
	CommitStatsRequest
	CommitStatsResponse
	CommitIsAncestorRequest
	CommitIsAncestorResponse
	TreeEntryRequest
	TreeEntryResponse
	CommitsBetweenRequest
	CommitsBetweenResponse
	CountCommitsRequest
	CountCommitsResponse
	TreeEntry
	GetTreeEntriesRequest
	GetTreeEntriesResponse
	ListFilesRequest
	ListFilesResponse
	FindCommitRequest
	FindCommitResponse
	ListCommitsByOidRequest
	ListCommitsByOidResponse
	FindAllCommitsRequest
	FindAllCommitsResponse
	FindCommitsRequest
	FindCommitsResponse
	CommitLanguagesRequest
	CommitLanguagesResponse
	RawBlameRequest
	RawBlameResponse
	LastCommitForPathRequest
	LastCommitForPathResponse
	CommitsByMessageRequest
	CommitsByMessageResponse
	FilterShasWithSignaturesRequest
	FilterShasWithSignaturesResponse
	ExtractCommitSignatureRequest
	ExtractCommitSignatureResponse
	ListConflictFilesRequest
	ConflictFileHeader
	ConflictFile
	ListConflictFilesResponse
	ResolveConflictsRequestHeader
	ResolveConflictsRequest
	ResolveConflictsResponse
	CommitDiffRequest
	CommitDiffResponse
	CommitDeltaRequest
	CommitDelta
	CommitDeltaResponse
	CommitPatchRequest
	CommitPatchResponse
	RawDiffRequest
	RawDiffResponse
	RawPatchRequest
	RawPatchResponse
	AddNamespaceRequest
	RemoveNamespaceRequest
	RenameNamespaceRequest
	NamespaceExistsRequest
	NamespaceExistsResponse
	AddNamespaceResponse
	RemoveNamespaceResponse
	RenameNamespaceResponse
	PostReceiveRequest
	PostReceiveResponse
	UserCreateBranchRequest
	UserCreateBranchResponse
	UserDeleteBranchRequest
	UserDeleteBranchResponse
	UserDeleteTagRequest
	UserDeleteTagResponse
	UserCreateTagRequest
	UserCreateTagResponse
	UserMergeBranchRequest
	UserMergeBranchResponse
	OperationBranchUpdate
	UserFFBranchRequest
	UserFFBranchResponse
	UserCherryPickRequest
	UserCherryPickResponse
	UserRevertRequest
	UserRevertResponse
	UserCommitFilesActionHeader
	UserCommitFilesAction
	UserCommitFilesRequestHeader
	UserCommitFilesRequest
	UserCommitFilesResponse
	UserRebaseRequest
	UserRebaseResponse
	UserSquashRequest
	UserSquashResponse
	FindDefaultBranchNameRequest
	FindDefaultBranchNameResponse
	FindAllBranchNamesRequest
	FindAllBranchNamesResponse
	FindAllTagNamesRequest
	FindAllTagNamesResponse
	FindRefNameRequest
	FindRefNameResponse
	FindLocalBranchesRequest
	FindLocalBranchesResponse
	FindLocalBranchResponse
	FindLocalBranchCommitAuthor
	FindAllBranchesRequest
	FindAllBranchesResponse
	FindAllTagsRequest
	FindAllTagsResponse
	RefExistsRequest
	RefExistsResponse
	CreateBranchRequest
	CreateBranchResponse
	DeleteBranchRequest
	DeleteBranchResponse
	FindBranchRequest
	FindBranchResponse
	DeleteRefsRequest
	DeleteRefsResponse
	ListBranchNamesContainingCommitRequest
	ListBranchNamesContainingCommitResponse
	ListTagNamesContainingCommitRequest
	ListTagNamesContainingCommitResponse
	AddRemoteRequest
	AddRemoteResponse
	RemoveRemoteRequest
	RemoveRemoteResponse
	FetchInternalRemoteRequest
	FetchInternalRemoteResponse
	UpdateRemoteMirrorRequest
	UpdateRemoteMirrorResponse
	RepositoryExistsRequest
	RepositoryExistsResponse
	RepositoryIsEmptyRequest
	RepositoryIsEmptyResponse
	RepackIncrementalRequest
	RepackIncrementalResponse
	RepackFullRequest
	RepackFullResponse
	GarbageCollectRequest
	GarbageCollectResponse
	RepositorySizeRequest
	RepositorySizeResponse
	ApplyGitattributesRequest
	ApplyGitattributesResponse
	FetchRemoteRequest
	FetchRemoteResponse
	CreateRepositoryRequest
	CreateRepositoryResponse
	GetArchiveRequest
	GetArchiveResponse
	HasLocalBranchesRequest
	HasLocalBranchesResponse
	FetchSourceBranchRequest
	FetchSourceBranchResponse
	FsckRequest
	FsckResponse
	WriteRefRequest
	WriteRefResponse
	FindMergeBaseRequest
	FindMergeBaseResponse
	CreateForkRequest
	CreateForkResponse
	IsRebaseInProgressRequest
	IsRebaseInProgressResponse
	IsSquashInProgressRequest
	IsSquashInProgressResponse
	CreateRepositoryFromURLRequest
	CreateRepositoryFromURLResponse
	CreateBundleRequest
	CreateBundleResponse
	WriteConfigRequest
	WriteConfigResponse
	CreateRepositoryFromBundleRequest
	CreateRepositoryFromBundleResponse
	ServerInfoRequest
	ServerInfoResponse
	Repository
	GitCommit
	CommitAuthor
	ExitStatus
	Branch
	Tag
	User
	InfoRefsRequest
	InfoRefsResponse
	PostUploadPackRequest
	PostUploadPackResponse
	PostReceivePackRequest
	PostReceivePackResponse
	SSHUploadPackRequest
	SSHUploadPackResponse
	SSHReceivePackRequest
	SSHReceivePackResponse
	WikiCommitDetails
	WikiPageVersion
	WikiPage
	WikiGetPageVersionsRequest
	WikiGetPageVersionsResponse
	WikiWritePageRequest
	WikiWritePageResponse
	WikiUpdatePageRequest
	WikiUpdatePageResponse
	WikiDeletePageRequest
	WikiDeletePageResponse
	WikiFindPageRequest
	WikiFindPageResponse
	WikiFindFileRequest
	WikiFindFileResponse
	WikiGetAllPagesRequest
	WikiGetAllPagesResponse
	WikiGetFormattedDataRequest
	WikiGetFormattedDataResponse
*/
package gitaly

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type GetBlobRequest struct {
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	// Object ID (SHA1) of the blob we want to get
	Oid string `protobuf:"bytes,2,opt,name=oid" json:"oid,omitempty"`
	// Maximum number of bytes we want to receive. Use '-1' to get the full blob no matter how big.
	Limit int64 `protobuf:"varint,3,opt,name=limit" json:"limit,omitempty"`
}

func (m *GetBlobRequest) Reset()                    { *m = GetBlobRequest{} }
func (m *GetBlobRequest) String() string            { return proto.CompactTextString(m) }
func (*GetBlobRequest) ProtoMessage()               {}
func (*GetBlobRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *GetBlobRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *GetBlobRequest) GetOid() string {
	if m != nil {
		return m.Oid
	}
	return ""
}

func (m *GetBlobRequest) GetLimit() int64 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type GetBlobResponse struct {
	// Blob size; present only in first response message
	Size int64 `protobuf:"varint,1,opt,name=size" json:"size,omitempty"`
	// Chunk of blob data
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	// Object ID of the actual blob returned. Empty if no blob was found.
	Oid string `protobuf:"bytes,3,opt,name=oid" json:"oid,omitempty"`
}

func (m *GetBlobResponse) Reset()                    { *m = GetBlobResponse{} }
func (m *GetBlobResponse) String() string            { return proto.CompactTextString(m) }
func (*GetBlobResponse) ProtoMessage()               {}
func (*GetBlobResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *GetBlobResponse) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *GetBlobResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *GetBlobResponse) GetOid() string {
	if m != nil {
		return m.Oid
	}
	return ""
}

type GetBlobsRequest struct {
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	// Revision/Path pairs of the blobs we want to get.
	RevisionPaths []*GetBlobsRequest_RevisionPath `protobuf:"bytes,2,rep,name=revision_paths,json=revisionPaths" json:"revision_paths,omitempty"`
	// Maximum number of bytes we want to receive. Use '-1' to get the full blobs no matter how big.
	Limit int64 `protobuf:"varint,3,opt,name=limit" json:"limit,omitempty"`
}

func (m *GetBlobsRequest) Reset()                    { *m = GetBlobsRequest{} }
func (m *GetBlobsRequest) String() string            { return proto.CompactTextString(m) }
func (*GetBlobsRequest) ProtoMessage()               {}
func (*GetBlobsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *GetBlobsRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *GetBlobsRequest) GetRevisionPaths() []*GetBlobsRequest_RevisionPath {
	if m != nil {
		return m.RevisionPaths
	}
	return nil
}

func (m *GetBlobsRequest) GetLimit() int64 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type GetBlobsRequest_RevisionPath struct {
	Revision string `protobuf:"bytes,1,opt,name=revision" json:"revision,omitempty"`
	Path     []byte `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
}

func (m *GetBlobsRequest_RevisionPath) Reset()                    { *m = GetBlobsRequest_RevisionPath{} }
func (m *GetBlobsRequest_RevisionPath) String() string            { return proto.CompactTextString(m) }
func (*GetBlobsRequest_RevisionPath) ProtoMessage()               {}
func (*GetBlobsRequest_RevisionPath) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

func (m *GetBlobsRequest_RevisionPath) GetRevision() string {
	if m != nil {
		return m.Revision
	}
	return ""
}

func (m *GetBlobsRequest_RevisionPath) GetPath() []byte {
	if m != nil {
		return m.Path
	}
	return nil
}

type GetBlobsResponse struct {
	// Blob size; present only on the first message per blob
	Size int64 `protobuf:"varint,1,opt,name=size" json:"size,omitempty"`
	// Chunk of blob data, could span over multiple messages.
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	// Object ID of the current blob. Only present on the first message per blob. Empty if no blob was found.
	Oid         string `protobuf:"bytes,3,opt,name=oid" json:"oid,omitempty"`
	IsSubmodule bool   `protobuf:"varint,4,opt,name=is_submodule,json=isSubmodule" json:"is_submodule,omitempty"`
	Mode        int32  `protobuf:"varint,5,opt,name=mode" json:"mode,omitempty"`
	Revision    string `protobuf:"bytes,6,opt,name=revision" json:"revision,omitempty"`
	Path        []byte `protobuf:"bytes,7,opt,name=path,proto3" json:"path,omitempty"`
}

func (m *GetBlobsResponse) Reset()                    { *m = GetBlobsResponse{} }
func (m *GetBlobsResponse) String() string            { return proto.CompactTextString(m) }
func (*GetBlobsResponse) ProtoMessage()               {}
func (*GetBlobsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *GetBlobsResponse) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *GetBlobsResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *GetBlobsResponse) GetOid() string {
	if m != nil {
		return m.Oid
	}
	return ""
}

func (m *GetBlobsResponse) GetIsSubmodule() bool {
	if m != nil {
		return m.IsSubmodule
	}
	return false
}

func (m *GetBlobsResponse) GetMode() int32 {
	if m != nil {
		return m.Mode
	}
	return 0
}

func (m *GetBlobsResponse) GetRevision() string {
	if m != nil {
		return m.Revision
	}
	return ""
}

func (m *GetBlobsResponse) GetPath() []byte {
	if m != nil {
		return m.Path
	}
	return nil
}

type LFSPointer struct {
	Size int64  `protobuf:"varint,1,opt,name=size" json:"size,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Oid  string `protobuf:"bytes,3,opt,name=oid" json:"oid,omitempty"`
}

func (m *LFSPointer) Reset()                    { *m = LFSPointer{} }
func (m *LFSPointer) String() string            { return proto.CompactTextString(m) }
func (*LFSPointer) ProtoMessage()               {}
func (*LFSPointer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *LFSPointer) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *LFSPointer) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *LFSPointer) GetOid() string {
	if m != nil {
		return m.Oid
	}
	return ""
}

type GetLFSPointersRequest struct {
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	BlobIds    []string    `protobuf:"bytes,2,rep,name=blob_ids,json=blobIds" json:"blob_ids,omitempty"`
}

func (m *GetLFSPointersRequest) Reset()                    { *m = GetLFSPointersRequest{} }
func (m *GetLFSPointersRequest) String() string            { return proto.CompactTextString(m) }
func (*GetLFSPointersRequest) ProtoMessage()               {}
func (*GetLFSPointersRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *GetLFSPointersRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *GetLFSPointersRequest) GetBlobIds() []string {
	if m != nil {
		return m.BlobIds
	}
	return nil
}

type GetLFSPointersResponse struct {
	LfsPointers []*LFSPointer `protobuf:"bytes,1,rep,name=lfs_pointers,json=lfsPointers" json:"lfs_pointers,omitempty"`
}

func (m *GetLFSPointersResponse) Reset()                    { *m = GetLFSPointersResponse{} }
func (m *GetLFSPointersResponse) String() string            { return proto.CompactTextString(m) }
func (*GetLFSPointersResponse) ProtoMessage()               {}
func (*GetLFSPointersResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GetLFSPointersResponse) GetLfsPointers() []*LFSPointer {
	if m != nil {
		return m.LfsPointers
	}
	return nil
}

func init() {
	proto.RegisterType((*GetBlobRequest)(nil), "gitaly.GetBlobRequest")
	proto.RegisterType((*GetBlobResponse)(nil), "gitaly.GetBlobResponse")
	proto.RegisterType((*GetBlobsRequest)(nil), "gitaly.GetBlobsRequest")
	proto.RegisterType((*GetBlobsRequest_RevisionPath)(nil), "gitaly.GetBlobsRequest.RevisionPath")
	proto.RegisterType((*GetBlobsResponse)(nil), "gitaly.GetBlobsResponse")
	proto.RegisterType((*LFSPointer)(nil), "gitaly.LFSPointer")
	proto.RegisterType((*GetLFSPointersRequest)(nil), "gitaly.GetLFSPointersRequest")
	proto.RegisterType((*GetLFSPointersResponse)(nil), "gitaly.GetLFSPointersResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for BlobService service

type BlobServiceClient interface {
	// GetBlob returns the contents of a blob object referenced by its object
	// ID. We use a stream to return a chunked arbitrarily large binary
	// response
	GetBlob(ctx context.Context, in *GetBlobRequest, opts ...grpc.CallOption) (BlobService_GetBlobClient, error)
	GetBlobs(ctx context.Context, in *GetBlobsRequest, opts ...grpc.CallOption) (BlobService_GetBlobsClient, error)
	GetLFSPointers(ctx context.Context, in *GetLFSPointersRequest, opts ...grpc.CallOption) (BlobService_GetLFSPointersClient, error)
}

type blobServiceClient struct {
	cc *grpc.ClientConn
}

func NewBlobServiceClient(cc *grpc.ClientConn) BlobServiceClient {
	return &blobServiceClient{cc}
}

func (c *blobServiceClient) GetBlob(ctx context.Context, in *GetBlobRequest, opts ...grpc.CallOption) (BlobService_GetBlobClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_BlobService_serviceDesc.Streams[0], c.cc, "/gitaly.BlobService/GetBlob", opts...)
	if err != nil {
		return nil, err
	}
	x := &blobServiceGetBlobClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BlobService_GetBlobClient interface {
	Recv() (*GetBlobResponse, error)
	grpc.ClientStream
}

type blobServiceGetBlobClient struct {
	grpc.ClientStream
}

func (x *blobServiceGetBlobClient) Recv() (*GetBlobResponse, error) {
	m := new(GetBlobResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *blobServiceClient) GetBlobs(ctx context.Context, in *GetBlobsRequest, opts ...grpc.CallOption) (BlobService_GetBlobsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_BlobService_serviceDesc.Streams[1], c.cc, "/gitaly.BlobService/GetBlobs", opts...)
	if err != nil {
		return nil, err
	}
	x := &blobServiceGetBlobsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BlobService_GetBlobsClient interface {
	Recv() (*GetBlobsResponse, error)
	grpc.ClientStream
}

type blobServiceGetBlobsClient struct {
	grpc.ClientStream
}

func (x *blobServiceGetBlobsClient) Recv() (*GetBlobsResponse, error) {
	m := new(GetBlobsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *blobServiceClient) GetLFSPointers(ctx context.Context, in *GetLFSPointersRequest, opts ...grpc.CallOption) (BlobService_GetLFSPointersClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_BlobService_serviceDesc.Streams[2], c.cc, "/gitaly.BlobService/GetLFSPointers", opts...)
	if err != nil {
		return nil, err
	}
	x := &blobServiceGetLFSPointersClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BlobService_GetLFSPointersClient interface {
	Recv() (*GetLFSPointersResponse, error)
	grpc.ClientStream
}

type blobServiceGetLFSPointersClient struct {
	grpc.ClientStream
}

func (x *blobServiceGetLFSPointersClient) Recv() (*GetLFSPointersResponse, error) {
	m := new(GetLFSPointersResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for BlobService service

type BlobServiceServer interface {
	// GetBlob returns the contents of a blob object referenced by its object
	// ID. We use a stream to return a chunked arbitrarily large binary
	// response
	GetBlob(*GetBlobRequest, BlobService_GetBlobServer) error
	GetBlobs(*GetBlobsRequest, BlobService_GetBlobsServer) error
	GetLFSPointers(*GetLFSPointersRequest, BlobService_GetLFSPointersServer) error
}

func RegisterBlobServiceServer(s *grpc.Server, srv BlobServiceServer) {
	s.RegisterService(&_BlobService_serviceDesc, srv)
}

func _BlobService_GetBlob_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetBlobRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlobServiceServer).GetBlob(m, &blobServiceGetBlobServer{stream})
}

type BlobService_GetBlobServer interface {
	Send(*GetBlobResponse) error
	grpc.ServerStream
}

type blobServiceGetBlobServer struct {
	grpc.ServerStream
}

func (x *blobServiceGetBlobServer) Send(m *GetBlobResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _BlobService_GetBlobs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetBlobsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlobServiceServer).GetBlobs(m, &blobServiceGetBlobsServer{stream})
}

type BlobService_GetBlobsServer interface {
	Send(*GetBlobsResponse) error
	grpc.ServerStream
}

type blobServiceGetBlobsServer struct {
	grpc.ServerStream
}

func (x *blobServiceGetBlobsServer) Send(m *GetBlobsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _BlobService_GetLFSPointers_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetLFSPointersRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlobServiceServer).GetLFSPointers(m, &blobServiceGetLFSPointersServer{stream})
}

type BlobService_GetLFSPointersServer interface {
	Send(*GetLFSPointersResponse) error
	grpc.ServerStream
}

type blobServiceGetLFSPointersServer struct {
	grpc.ServerStream
}

func (x *blobServiceGetLFSPointersServer) Send(m *GetLFSPointersResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _BlobService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.BlobService",
	HandlerType: (*BlobServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetBlob",
			Handler:       _BlobService_GetBlob_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetBlobs",
			Handler:       _BlobService_GetBlobs_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetLFSPointers",
			Handler:       _BlobService_GetLFSPointers_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "blob.proto",
}

func init() { proto.RegisterFile("blob.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 451 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xcd, 0x8e, 0xd3, 0x30,
	0x10, 0xc6, 0x9b, 0xed, 0xdf, 0x24, 0x2c, 0x2b, 0x0b, 0x76, 0x43, 0x24, 0x50, 0x88, 0x38, 0xe4,
	0x54, 0xa1, 0x22, 0xae, 0x2b, 0xc1, 0x61, 0x57, 0x68, 0x91, 0x58, 0xb9, 0x0f, 0x50, 0x25, 0xc4,
	0xa5, 0x96, 0xdc, 0x3a, 0x78, 0xdc, 0x95, 0xca, 0x6b, 0xf1, 0x4c, 0x48, 0x3c, 0x06, 0xb2, 0xf3,
	0xd3, 0xd0, 0xd2, 0x53, 0x6e, 0x33, 0x63, 0x7f, 0x3f, 0xe3, 0x99, 0x04, 0x20, 0x97, 0x2a, 0x9f,
	0x96, 0x5a, 0x19, 0x45, 0x87, 0xdf, 0x85, 0xc9, 0xe4, 0x2e, 0x0a, 0x70, 0x95, 0x69, 0x5e, 0x54,
	0xd5, 0x44, 0xc2, 0xc5, 0x1d, 0x37, 0x9f, 0xa4, 0xca, 0x19, 0xff, 0xb1, 0xe5, 0x68, 0xe8, 0x0c,
	0x40, 0xf3, 0x52, 0xa1, 0x30, 0x4a, 0xef, 0x42, 0x12, 0x93, 0xd4, 0x9f, 0xd1, 0x69, 0x05, 0x9e,
	0xb2, 0xf6, 0x84, 0x75, 0x6e, 0xd1, 0x4b, 0xf0, 0x94, 0x28, 0xc2, 0xb3, 0x98, 0xa4, 0x13, 0x66,
	0x43, 0xfa, 0x1c, 0x06, 0x52, 0xac, 0x85, 0x09, 0xbd, 0x98, 0xa4, 0x1e, 0xab, 0x92, 0xe4, 0x1e,
	0x9e, 0xb5, 0x6a, 0x58, 0xaa, 0x0d, 0x72, 0x4a, 0xe1, 0x1c, 0xc5, 0x4f, 0xee, 0x84, 0x3c, 0xe6,
	0x62, 0x5b, 0x2b, 0x32, 0x93, 0x39, 0xbe, 0x80, 0xb9, 0xb8, 0x91, 0xf0, 0x5a, 0x89, 0xe4, 0x0f,
	0x69, 0xd9, 0xb0, 0x8f, 0xf9, 0x7b, 0xb8, 0xd0, 0xfc, 0x51, 0xa0, 0x50, 0x9b, 0x45, 0x99, 0x99,
	0x15, 0x86, 0x67, 0xb1, 0x97, 0xfa, 0xb3, 0xb7, 0x0d, 0xee, 0x40, 0x64, 0xca, 0xea, 0xdb, 0x0f,
	0x99, 0x59, 0xb1, 0xa7, 0xba, 0x93, 0xe1, 0xff, 0xfb, 0x8e, 0x6e, 0x20, 0xe8, 0x82, 0x68, 0x04,
	0xe3, 0x06, 0xe6, 0x4c, 0x4e, 0x58, 0x9b, 0xdb, 0xe6, 0xad, 0x8b, 0xa6, 0x79, 0x1b, 0x27, 0xbf,
	0x08, 0x5c, 0xee, 0x5d, 0xf4, 0x7d, 0x39, 0xfa, 0x06, 0x02, 0x81, 0x0b, 0xdc, 0xe6, 0x6b, 0x55,
	0x6c, 0x25, 0x0f, 0xcf, 0x63, 0x92, 0x8e, 0x99, 0x2f, 0x70, 0xde, 0x94, 0x2c, 0xd1, 0x5a, 0x15,
	0x3c, 0x1c, 0xc4, 0x24, 0x1d, 0x30, 0x17, 0xff, 0xe3, 0x7a, 0x78, 0xc2, 0xf5, 0xa8, 0xe3, 0xfa,
	0x16, 0xe0, 0xcb, 0xed, 0xfc, 0x41, 0x89, 0x8d, 0xe1, 0xba, 0xc7, 0xa0, 0x97, 0xf0, 0xe2, 0x8e,
	0x9b, 0x3d, 0x55, 0xaf, 0x69, 0xbf, 0x84, 0xb1, 0xfd, 0x28, 0x16, 0xa2, 0xa8, 0xe6, 0x3c, 0x61,
	0x23, 0x9b, 0x7f, 0x2e, 0x30, 0xf9, 0x0a, 0x57, 0x87, 0x3a, 0xf5, 0x53, 0x7f, 0x80, 0x40, 0x2e,
	0x71, 0x51, 0xd6, 0xf5, 0x90, 0xb8, 0x05, 0x69, 0xa5, 0xf6, 0x10, 0xe6, 0xcb, 0x25, 0x36, 0xf0,
	0xd9, 0x6f, 0x02, 0xbe, 0x9d, 0xd9, 0x9c, 0xeb, 0x47, 0xf1, 0x8d, 0xd3, 0x1b, 0x18, 0xd5, 0x53,
	0xa4, 0x57, 0x07, 0xcb, 0x55, 0xb7, 0x14, 0x5d, 0x1f, 0xd5, 0x2b, 0x0b, 0xc9, 0x93, 0x77, 0x84,
	0x7e, 0x84, 0x71, 0xb3, 0x05, 0xf4, 0xfa, 0xc4, 0x76, 0x46, 0xe1, 0xf1, 0x41, 0x87, 0x62, 0xee,
	0xbe, 0xf7, 0x4e, 0x8f, 0xf4, 0x55, 0xe7, 0xfe, 0xf1, 0x1b, 0x47, 0xaf, 0x4f, 0x1d, 0xef, 0x49,
	0xf3, 0xa1, 0xfb, 0x97, 0xbc, 0xff, 0x1b, 0x00, 0x00, 0xff, 0xff, 0xac, 0x26, 0x0e, 0x73, 0x6f,
	0x04, 0x00, 0x00,
}
