# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: praefect.proto

require 'google/protobuf'

require 'lint_pb'
require 'shared_pb'
Google::Protobuf::DescriptorPool.generated_pool.build do
  add_file("praefect.proto", :syntax => :proto3) do
    add_message "gitaly.GetRepositoryMetadataRequest" do
      oneof :query do
        optional :repository_id, :int64, 1
        optional :path, :message, 2, "gitaly.GetRepositoryMetadataRequest.Path"
      end
    end
    add_message "gitaly.GetRepositoryMetadataRequest.Path" do
      optional :virtual_storage, :string, 1
      optional :relative_path, :string, 2
    end
    add_message "gitaly.GetRepositoryMetadataResponse" do
      optional :repository_id, :int64, 1
      optional :virtual_storage, :string, 2
      optional :relative_path, :string, 3
      optional :replica_path, :string, 4
      optional :primary, :string, 5
      optional :generation, :int64, 6
      repeated :replicas, :message, 7, "gitaly.GetRepositoryMetadataResponse.Replica"
    end
    add_message "gitaly.GetRepositoryMetadataResponse.Replica" do
      optional :storage, :string, 1
      optional :assigned, :bool, 2
      optional :generation, :int64, 4
      optional :healthy, :bool, 5
      optional :valid_primary, :bool, 6
    end
    add_message "gitaly.SetReplicationFactorRequest" do
      optional :virtual_storage, :string, 1
      optional :relative_path, :string, 2
      optional :replication_factor, :int32, 3
    end
    add_message "gitaly.SetReplicationFactorResponse" do
      repeated :storages, :string, 1
    end
    add_message "gitaly.SetAuthoritativeStorageRequest" do
      optional :virtual_storage, :string, 1
      optional :relative_path, :string, 2
      optional :authoritative_storage, :string, 3
    end
    add_message "gitaly.SetAuthoritativeStorageResponse" do
    end
    add_message "gitaly.DatalossCheckRequest" do
      optional :virtual_storage, :string, 1
      optional :include_partially_replicated, :bool, 2
    end
    add_message "gitaly.DatalossCheckResponse" do
      repeated :repositories, :message, 2, "gitaly.DatalossCheckResponse.Repository"
    end
    add_message "gitaly.DatalossCheckResponse.Repository" do
      optional :relative_path, :string, 1
      repeated :storages, :message, 2, "gitaly.DatalossCheckResponse.Repository.Storage"
      optional :unavailable, :bool, 3
      optional :primary, :string, 4
    end
    add_message "gitaly.DatalossCheckResponse.Repository.Storage" do
      optional :name, :string, 1
      optional :behind_by, :int64, 2
      optional :assigned, :bool, 3
      optional :healthy, :bool, 4
      optional :valid_primary, :bool, 5
    end
    add_message "gitaly.RepositoryReplicasRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
    end
    add_message "gitaly.RepositoryReplicasResponse" do
      optional :primary, :message, 1, "gitaly.RepositoryReplicasResponse.RepositoryDetails"
      repeated :replicas, :message, 2, "gitaly.RepositoryReplicasResponse.RepositoryDetails"
    end
    add_message "gitaly.RepositoryReplicasResponse.RepositoryDetails" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :checksum, :string, 2
    end
  end
end

module Gitaly
  GetRepositoryMetadataRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.GetRepositoryMetadataRequest").msgclass
  GetRepositoryMetadataRequest::Path = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.GetRepositoryMetadataRequest.Path").msgclass
  GetRepositoryMetadataResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.GetRepositoryMetadataResponse").msgclass
  GetRepositoryMetadataResponse::Replica = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.GetRepositoryMetadataResponse.Replica").msgclass
  SetReplicationFactorRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.SetReplicationFactorRequest").msgclass
  SetReplicationFactorResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.SetReplicationFactorResponse").msgclass
  SetAuthoritativeStorageRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.SetAuthoritativeStorageRequest").msgclass
  SetAuthoritativeStorageResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.SetAuthoritativeStorageResponse").msgclass
  DatalossCheckRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.DatalossCheckRequest").msgclass
  DatalossCheckResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.DatalossCheckResponse").msgclass
  DatalossCheckResponse::Repository = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.DatalossCheckResponse.Repository").msgclass
  DatalossCheckResponse::Repository::Storage = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.DatalossCheckResponse.Repository.Storage").msgclass
  RepositoryReplicasRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.RepositoryReplicasRequest").msgclass
  RepositoryReplicasResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.RepositoryReplicasResponse").msgclass
  RepositoryReplicasResponse::RepositoryDetails = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.RepositoryReplicasResponse.RepositoryDetails").msgclass
end
