# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: errors.proto

require 'google/protobuf'

Google::Protobuf::DescriptorPool.generated_pool.build do
  add_file("errors.proto", :syntax => :proto3) do
    add_message "gitaly.AccessCheckError" do
      optional :error_message, :string, 1
      optional :protocol, :string, 2
      optional :user_id, :string, 3
      optional :changes, :bytes, 4
    end
  end
end

module Gitaly
  AccessCheckError = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.AccessCheckError").msgclass
end