#include <iostream>
#include <string>
#include <cstdint>

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
using zksnark::VerifyReply;
using zksnark::VerifyRequest;
using zksnark::Zksnark;

ZCJoinSplit *js;

typedef std::array<libzcash::JSInput, ZC_NUM_JS_INPUTS> ProveInputs;
typedef std::array<libzcash::SproutNote, ZC_NUM_JS_OUTPUTS> ProveOutnotes;

typedef std::array<uint256, ZC_NUM_JS_INPUTS> NullifierArray;
typedef std::array<uint256, ZC_NUM_JS_OUTPUTS> CommitmentArray;

bool string_to_uint256(const string &data, uint256 &result)
{
    if (data.size() != 32)
        return false;
    const unsigned char *data_mem = (const unsigned char *)data.c_str();
    std::vector<unsigned char> data_vec(data_mem, data_mem + data.size());
    result = uint256(data_vec);
    return true;
}

bool string_to_uint252(const string &data, uint252 &result)
{
    uint256 data256;
    bool success = string_to_uint256(data, data256);
    if (!success || *(data256.end() - 1) & 0xF0) // TODO: check for endianness
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

bool convert_input_note(const zksnark::Note &zk_note, libzcash::SproutNote &note)
{
    note.value_ = zk_note.value();
    bool success = true;
    success &= string_to_uint256(zk_note.cm(), note.cm);
    success &= string_to_uint256(zk_note.r(), note.r);
    success &= string_to_uint256(zk_note.rho(), note.rho);
    success &= string_to_uint256(zk_note.apk(), note.a_pk);
    success &= string_to_uint256(zk_note.nf(), note.nf);
    return success;
}

bool convert_output_note(const zksnark::Note &zk_note, libzcash::SproutNote &note)
{
    note.value_ = zk_note.value();
    bool success = true;
    success &= string_to_uint256(zk_note.cm(), note.cm);
    success &= string_to_uint256(zk_note.r(), note.r);
    success &= string_to_uint256(zk_note.rho(), note.rho);
    success &= string_to_uint256(zk_note.apk(), note.a_pk);
    return success;
}


bool transform_prove_request(const ProveRequest *request,
                             ProveInputs &inputs,
                             ProveOutnotes &out_notes,
                             uint256 &hsig,
                             uint252 &phi,
                             uint256 &rt,
                             uint64_t &reward)
{
    if (request->inputs_size() != ZC_NUM_JS_INPUTS || request->outnotes_size() != ZC_NUM_JS_OUTPUTS)
        return false;

    // Convert inputs
    bool success = true;
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
        cout << "input.witnesspath().authpath().size():" << input.witnesspath().authpath().size() << '\n';
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
        cout << "input.witnesspath().index().size():" << input.witnesspath().index().size() << '\n';

        // The length of the witness and merkle tree's depth must match
        if (auth_path.size() != index.size() || index.size() != INCREMENTAL_MERKLE_TREE_DEPTH)
            return false;

        inputs[i].witness = ZCIncrementalWitness(auth_path, index);

        // Convert note
        success &= convert_input_note(input.note(), inputs[i].note);
        cout << "inpnote: \n";
        cout << "key: " << inputs[i].key.inner().GetHex() << '\n';
        cout << "path_length: " << inputs[i].witness.path().index.size() << '\n';
        cout << "note apk: " << inputs[i].note.a_pk.GetHex() << '\n';
        cout << "note nf: " << inputs[i].note.nf.GetHex() << '\n';
        cout << "note rho: " << inputs[i].note.rho.GetHex() << '\n';
        cout << '\n';

        i++;
    }
    cout << "success after input: " << success << '\n';

    // Convert outnotes
    i = 0;
    for (auto &outnote : request->outnotes())
    {
        success &= convert_output_note(outnote, out_notes[i]);
        cout << "outnote: \n";
        cout << "apk: " << out_notes[i].a_pk.GetHex() << '\n';
        cout << "value: " << out_notes[i].value() << '\n';
        cout << "rho: " << out_notes[i].rho.GetHex() << '\n';
        cout << "r: " << out_notes[i].r.GetHex() << '\n';
        cout << "cm: " << out_notes[i].cm.GetHex() << '\n';
        cout << '\n';
        i++;
    }
    cout << "success after output: " << success << '\n';

    // Convert hsig
    success &= string_to_uint256(request->hsig(), hsig);
    cout << "hsig: " << hsig.GetHex() << '\n';

    // Convert phi
    success &= string_to_uint252(request->phi(), phi);
    cout << "phi: " << phi.inner().GetHex() << '\n';

    // Convert rt
    success &= string_to_uint256(request->rt(), rt);
    cout << "rt: " << rt.GetHex() << '\n';

    // Convert reward
    reward = request->reward();
    cout << "reward: " << reward << '\n';

    return success;
}

int print_proof(libzcash::PHGRProof &proof) {
    printf("g_A: ");
    proof.g_A.print();
    printf("g_A_prime: ");
    proof.g_A_prime.print();
    printf("g_B: ");
    proof.g_B.print();
    printf("g_B_prime: ");
    proof.g_B_prime.print();
    printf("g_C: ");
    proof.g_C.print();
    printf("g_C_prime: ");
    proof.g_C_prime.print();
    printf("g_K: ");
    proof.g_K.print();
    printf("g_H: ");
    proof.g_H.print();

    return 0;
}

bool transform_prove_reply(libzcash::PHGRProof &proof, zksnark::PHGRProof &zk_proof)
{
    zk_proof.set_g_a(proof.g_A.to_string());
    zk_proof.set_g_a_prime(proof.g_A_prime.to_string());
    zk_proof.set_g_b(proof.g_B.to_string());
    zk_proof.set_g_b_prime(proof.g_B_prime.to_string());
    zk_proof.set_g_c(proof.g_C.to_string());
    zk_proof.set_g_c_prime(proof.g_C_prime.to_string());
    zk_proof.set_g_h(proof.g_H.to_string());
    zk_proof.set_g_k(proof.g_K.to_string());
    return true;
}

bool transform_verify_request(const VerifyRequest *request,
                              libzcash::PHGRProof &proof,
                              NullifierArray &nullifiers,
                              CommitmentArray &commitments,
                              uint256 &hsig,
                              uint256 &rt,
                              uint64_t &reward)
{
    if (request->nullifiers_size() != ZC_NUM_JS_INPUTS || request->commits_size() != ZC_NUM_JS_OUTPUTS)
        return false;

    cout << "Done check size\n";
    // Convert PHGRProof
    printf("Checking: g_A\n");
    bool success = proof.g_A.from_string(request->proof().g_a());
    printf("Checking: g_A_prime\n");
    success &= proof.g_A_prime.from_string(request->proof().g_a_prime());
    printf("Checking: g_B\n");
    success &= proof.g_B.from_string(request->proof().g_b());
    printf("Checking: g_B_prime\n");
    success &= proof.g_B_prime.from_string(request->proof().g_b_prime());
    printf("Checking: g_C\n");
    success &= proof.g_C.from_string(request->proof().g_c());
    printf("Checking: g_C_prime\n");
    success &= proof.g_C_prime.from_string(request->proof().g_c_prime());
    printf("Checking: g_K\n");
    success &= proof.g_K.from_string(request->proof().g_k());
    printf("Checking: g_H\n");
    success &= proof.g_H.from_string(request->proof().g_h());
    cout << "convert PHGRProof: " << success << '\n';

    // Convert nullifiers
    for (int i = 0; i < request->nullifiers_size(); ++i)
    {
        auto nf = request->nullifiers(i);
        string_to_uint256(nf, nullifiers[i]);
    }

    // Convert commits
    for (int i = 0; i < request->commits_size(); ++i)
    {
        auto cm = request->commits(i);
        string_to_uint256(cm, commitments[i]);
    }

    // Convert hsig
    success &= string_to_uint256(request->hsig(), hsig);
    cout << "hsig: " << hsig.GetHex() << '\n';

    // Convert rt
    success &= string_to_uint256(request->rt(), rt);
    cout << "rt: " << rt.GetHex() << '\n';

    // Convert reward
    reward = request->reward();

    return success;
}

class ZksnarkImpl final : public Zksnark::Service
{
    Status Prove(ServerContext *context, const ProveRequest *request, ProveReply *reply) override
    {
        cout << "Starting Prove, request->inputs_size(): " << request->inputs_size() << '\n';
        ProveInputs inputs;
        ProveOutnotes out_notes;
        uint256 hsig, rt;
        uint252 phi;
        uint64_t reward;
        bool success = transform_prove_request(request, inputs, out_notes, hsig, phi, rt, reward);
        cout << "transform_prove_request status: " << success << '\n';

        if (!success)
            return Status::OK;

        bool compute_proof = true;
        uint64_t vpub_old = reward;
        auto proof = js->prove(inputs, out_notes, vpub_old, rt, hsig, phi, compute_proof);
        print_proof(proof);

        zksnark::PHGRProof *zk_proof = new zksnark::PHGRProof();
        success = transform_prove_reply(proof, *zk_proof);
        cout << "transform_prove_reply status: " << success << '\n';
        cout << "setting allocated_proof\n";
        reply->set_allocated_proof(zk_proof);
        return Status::OK;
    }

    Status Verify(ServerContext *context, const VerifyRequest *request, VerifyReply *reply) override
    {
        libzcash::PHGRProof proof;
        uint256 hsig, rt;
        NullifierArray nullifiers;
        CommitmentArray commitments;
        NullifierArray macs; // TODO(@0xbunyip): add mac to successfully run js->verify
        uint64_t reward;
        bool success = transform_verify_request(request, proof, nullifiers, commitments, hsig, rt, reward);
        cout << "transform_verify_request status: " << success << '\n';

        uint64_t vpub_old = reward;
        bool valid = false;
        if (success)
            valid = js->verify(proof, macs, nullifiers, commitments, vpub_old, rt, hsig);
        reply->set_valid(valid);
        return Status::OK;
    }
};

void RunServer()
{
    // Creating zksnark circuit and load params
    js = ZCJoinSplit::Prepared("/home/ubuntu/go/src/github.com/ninjadotorg/cash-prototype/privacy/server/build/verifying.key",
                               "/home/ubuntu/go/src/github.com/ninjadotorg/cash-prototype/privacy/server/build/proving.key");
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
