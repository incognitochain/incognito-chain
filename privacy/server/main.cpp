#include <iostream>
#include <string>

#include <grpc/grpc.h>
#include <grpcpp/server.h>
#include <grpcpp/server_builder.h>
#include <grpcpp/server_context.h>

#include "JoinSplit.hpp"
#include "Address.hpp"
#include "../proto/zksnark.grpc.pb.h"

using namespace std;
using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::Status;
using zksnark::ProveReply;
using zksnark::ProveRequest;
using zksnark::Zksnark;

ZCJoinSplit *js;

typedef std::array<libzcash::JSInput, ZC_NUM_JS_INPUTS> ProveInputs;
typedef std::array<libzcash::SproutNote, ZC_NUM_JS_OUTPUTS> ProveOutnotes;


bool string_to_uint256(const string &data, uint256 &result)
{
    const unsigned char *data_mem = (const unsigned char *)data.c_str();
    std::vector<unsigned char> data_vec(data_mem, data_mem + data.size());
    result = uint256(data_vec);
    return true;
}

bool string_to_uint252(const string &data, uint252 &result)
{
    uint256 data256;
    bool success = string_to_uint256(data, data256);
    if (!success || *data256.begin() & 0xF0) // TODO: check for endianess
        return false;
    result = uint252(data256);
    return true;
}

bool string_to_bools(const string &data, vector<bool> &result)
{
    const unsigned char *data_mem = (const unsigned char *)data.c_str();
    result.resize(data.size() * 8);
    for (int i = 0; i < data.size(); ++i)
        for (int j = 0; j < 8; ++j)
            result[i * 8 + j] = bool((data_mem[i] >> j) & 1);
    return true;
}

bool convert_note(const zksnark::Note &zk_note, libzcash::SproutNote &note)
{
    note.value_ = zk_note.value();
    string_to_uint256(zk_note.cm(), note.cm);
    string_to_uint256(zk_note.r(), note.r);
    string_to_uint256(zk_note.rho(), note.rho);
    string_to_uint256(zk_note.apk(), note.a_pk);
    string_to_uint256(zk_note.nf(), note.nf);
    return true;
}

bool transform_prove_request(const ProveRequest *request,
                             ProveInputs &inputs,
                             ProveOutnotes &out_notes,
                             uint256 &hsig, uint252 &phi, uint256 &rt)
{
    if (request->inputs_size() != ZC_NUM_JS_INPUTS || request->outnotes_size() != ZC_NUM_JS_OUTPUTS)
        return false;

    // Convert inputs
    int i = 0;
    for (auto &input : request->inputs())
    {
        // Convert spending key
        uint252 key;
        bool s = string_to_uint252(input.spendingkey(), key);
        if (!s)
            return false;
        inputs[i].key = libzcash::SproutSpendingKey(key);

        // Convert witness
        // Witness' authentication path
        vector<vector<bool>> auth_path;
        auto auth_path_strs = input.witnesspath().authpath();
        for (auto &path_str : auth_path_strs)
        {
            vector<bool> path;
            s &= string_to_bools(path_str.hash(), path);
            if (!s)
                return false;
            auth_path.push_back(path);
        }

        // Witness' index
        vector<bool> index;
        for (auto &idx : input.witnesspath().index())
        {
            index.push_back(idx);
        }

        // The length of the witness and merkle tree's depth must match
        if (auth_path.size() != index.size() || index.size() != INCREMENTAL_MERKLE_TREE_DEPTH)
            return false;

        inputs[i].witness = ZCIncrementalWitness(auth_path, index);

        // Convert note
        convert_note(input.note(), inputs[i].note);
        i++;
    }

    // Convert outnotes
    i = 0;
    for (auto &outnote : request->outnotes())
    {
        convert_note(outnote, out_notes[i]);
        i++;
    }

    // Convert hsig
    bool success = true;
    success &= string_to_uint256(request->hsig(), hsig);

    // Convert phi
    success &= string_to_uint252(request->phi(), phi);

    // Convert rt
    success &= string_to_uint256(request->rt(), rt);

    return success;
}

class ZksnarkImpl final : public Zksnark::Service
{
    Status Prove(ServerContext *context, const ProveRequest *request, ProveReply *reply) override
    {
        cout << request->inputs_size() << '\n';
        ProveInputs inputs;
        ProveOutnotes out_notes;
        uint256 hsig, rt;
        uint252 phi;
        bool success = transform_prove_request(request, inputs, out_notes, hsig, phi, rt);
        cout << success << '\n';
        return Status::OK;
    }
};

void RunServer()
{
    // Creating zksnark circuit and load params
    js = ZCJoinSplit::Prepared("/home/ubuntu/go/src/github.com/thaibaoautonomous/btcd/privacy/server/build/verifying.key",
                               "/home/ubuntu/go/src/github.com/thaibaoautonomous/btcd/privacy/server/build/proving.key");
    cout << "Done preparing zksnark\n";

    // Run server
    string server_address("0.0.0.0:50052");
    ZksnarkImpl service;

    ServerBuilder builder;
    builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
    builder.RegisterService(&service);
    unique_ptr<Server> server(builder.BuildAndStart());
    cout << "Listening on: " << server_address << '\n';
    server->Wait();
}

int main(int argc, char const *argv[])
{
    RunServer();
    return 0;
}
