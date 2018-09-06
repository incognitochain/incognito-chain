// Generated by the gRPC C++ plugin.
// If you make any local change, they will be lost.
// source: zksnark.proto

#include "zksnark.pb.h"
#include "zksnark.grpc.pb.h"

#include <grpcpp/impl/codegen/async_stream.h>
#include <grpcpp/impl/codegen/async_unary_call.h>
#include <grpcpp/impl/codegen/channel_interface.h>
#include <grpcpp/impl/codegen/client_unary_call.h>
#include <grpcpp/impl/codegen/method_handler_impl.h>
#include <grpcpp/impl/codegen/rpc_service_method.h>
#include <grpcpp/impl/codegen/service_type.h>
#include <grpcpp/impl/codegen/sync_stream.h>
namespace zksnark {

static const char* Zksnark_method_names[] = {
  "/zksnark.Zksnark/Prove",
  "/zksnark.Zksnark/Verify",
};

std::unique_ptr< Zksnark::Stub> Zksnark::NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options) {
  (void)options;
  std::unique_ptr< Zksnark::Stub> stub(new Zksnark::Stub(channel));
  return stub;
}

Zksnark::Stub::Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel)
  : channel_(channel), rpcmethod_Prove_(Zksnark_method_names[0], ::grpc::internal::RpcMethod::NORMAL_RPC, channel)
  , rpcmethod_Verify_(Zksnark_method_names[1], ::grpc::internal::RpcMethod::NORMAL_RPC, channel)
  {}

::grpc::Status Zksnark::Stub::Prove(::grpc::ClientContext* context, const ::zksnark::ProveRequest& request, ::zksnark::ProveReply* response) {
  return ::grpc::internal::BlockingUnaryCall(channel_.get(), rpcmethod_Prove_, context, request, response);
}

::grpc::ClientAsyncResponseReader< ::zksnark::ProveReply>* Zksnark::Stub::AsyncProveRaw(::grpc::ClientContext* context, const ::zksnark::ProveRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::zksnark::ProveReply>::Create(channel_.get(), cq, rpcmethod_Prove_, context, request, true);
}

::grpc::ClientAsyncResponseReader< ::zksnark::ProveReply>* Zksnark::Stub::PrepareAsyncProveRaw(::grpc::ClientContext* context, const ::zksnark::ProveRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::zksnark::ProveReply>::Create(channel_.get(), cq, rpcmethod_Prove_, context, request, false);
}

::grpc::Status Zksnark::Stub::Verify(::grpc::ClientContext* context, const ::zksnark::VerifyRequest& request, ::zksnark::VerifyReply* response) {
  return ::grpc::internal::BlockingUnaryCall(channel_.get(), rpcmethod_Verify_, context, request, response);
}

::grpc::ClientAsyncResponseReader< ::zksnark::VerifyReply>* Zksnark::Stub::AsyncVerifyRaw(::grpc::ClientContext* context, const ::zksnark::VerifyRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::zksnark::VerifyReply>::Create(channel_.get(), cq, rpcmethod_Verify_, context, request, true);
}

::grpc::ClientAsyncResponseReader< ::zksnark::VerifyReply>* Zksnark::Stub::PrepareAsyncVerifyRaw(::grpc::ClientContext* context, const ::zksnark::VerifyRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::zksnark::VerifyReply>::Create(channel_.get(), cq, rpcmethod_Verify_, context, request, false);
}

Zksnark::Service::Service() {
  AddMethod(new ::grpc::internal::RpcServiceMethod(
      Zksnark_method_names[0],
      ::grpc::internal::RpcMethod::NORMAL_RPC,
      new ::grpc::internal::RpcMethodHandler< Zksnark::Service, ::zksnark::ProveRequest, ::zksnark::ProveReply>(
          std::mem_fn(&Zksnark::Service::Prove), this)));
  AddMethod(new ::grpc::internal::RpcServiceMethod(
      Zksnark_method_names[1],
      ::grpc::internal::RpcMethod::NORMAL_RPC,
      new ::grpc::internal::RpcMethodHandler< Zksnark::Service, ::zksnark::VerifyRequest, ::zksnark::VerifyReply>(
          std::mem_fn(&Zksnark::Service::Verify), this)));
}

Zksnark::Service::~Service() {
}

::grpc::Status Zksnark::Service::Prove(::grpc::ServerContext* context, const ::zksnark::ProveRequest* request, ::zksnark::ProveReply* response) {
  (void) context;
  (void) request;
  (void) response;
  return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
}

::grpc::Status Zksnark::Service::Verify(::grpc::ServerContext* context, const ::zksnark::VerifyRequest* request, ::zksnark::VerifyReply* response) {
  (void) context;
  (void) request;
  (void) response;
  return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
}


}  // namespace zksnark

