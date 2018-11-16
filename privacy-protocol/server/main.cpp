#include <iostream>
#include <string>
#include <cstdint>
#include <unistd.h>

#include <grpc/grpc.h>
#include <grpcpp/server.h>
#include <grpcpp/server_builder.h>
#include <grpcpp/server_context.h>

#include "IncrementalMerkleTree.hpp"

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
typedef std::array<uint256, ZC_NUM_JS_INPUTS> MacArray;
typedef std::array<uint256, ZC_NUM_JS_OUTPUTS> CommitmentArray;
typedef std::array<uint256, ZC_NUM_JS_INPUTS> RtArray;

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

bool path_string_to_bools(const string &data, vector<bool> &result)
{
    const unsigned char *data_mem = (const unsigned char *)data.c_str();
    result.resize(data.size() * 8);
    for (int i = 0; i < data.size(); ++i)
        for (int j = 0; j < 8; ++j)
            result[i * 8 + j] = bool((data_mem[i] >> (7 - j)) & 1); // 7-j since each byte must be convert to MSB
    return true;
}

bool convert_input_note(const zksnark::Note &zk_note, libzcash::SproutNote &note)
{
    note.value_ = zk_note.value();
    bool success = true;
    success &= string_to_uint256(zk_note.cm(), note.cm);
    if (!success)
        cout << "convert_input_note: fail to convert cm\n";
    success &= string_to_uint256(zk_note.r(), note.r);
    if (!success)
        cout << "convert_input_note: fail to convert r\n";
    success &= string_to_uint256(zk_note.rho(), note.rho);
    if (!success)
        cout << "convert_input_note: fail to convert rho\n";
    success &= string_to_uint256(zk_note.apk(), note.a_pk);
    if (!success)
        cout << "convert_input_note: fail to convert apk\n";
    success &= string_to_uint256(zk_note.nf(), note.nf);
    if (!success)
        cout << "convert_input_note: fail to convert nf\n";
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

bool convert_witness(const zksnark::MerklePath &paths, ZCIncrementalWitness &witness)
{
    vector<vector<bool>> auth_path;
    for (auto &path_hash : paths.authpath())
    {
        vector<bool> path;
        bool s = path_string_to_bools(path_hash.hash(), path);
        if (!s)
            return false;
        auth_path.push_back(path);
    }

    // Witness' index
    vector<bool> index;
    for (auto &idx : paths.index())
        index.push_back(idx);

    // The length of the witness and merkle tree's depth must match
    if (auth_path.size() != index.size() || index.size() != INCREMENTAL_MERKLE_TREE_DEPTH)
        return false;

    reverse(auth_path.begin(), auth_path.end()); // Reverse path for libsnark
    reverse(index.begin(), index.end());
    witness = ZCIncrementalWitness(auth_path, index);
    return true;
}

bool transform_prove_request(const ProveRequest *request,
                             ProveInputs &inputs,
                             ProveOutnotes &out_notes,
                             uint256 &hsig,
                             uint252 &phi,
                             RtArray &rts,
                             uint64_t &reward,
                             uint64_t &fee,
                             unsigned char &address_last_byte)
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
        ZCIncrementalWitness witness;
        s = convert_witness(input.witnesspath(), witness);
        if (!s)
            return false;

        inputs[i].witness = witness;

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

    // Convert phi
    success &= string_to_uint252(request->phi(), phi);

    // Convert rt
    i = 0;
    for (auto &rt : request->rts())
    {
        success &= string_to_uint256(rt, rts[i++]);
    }

    // Convert reward
    reward = request->reward();

    // Convert fee
    fee = request->fee();

    // Convert address last byte
    address_last_byte = request->addresslastbyte() & ((1 << 8) - 1);

    return success;
}

int print_proof(libzcash::PHGRProof &proof)
{
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
                              MacArray &macs,
                              uint256 &hsig,
                              RtArray &rts,
                              uint64_t &reward,
                              uint64_t &fee,
                              unsigned char &address_last_byte)
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

    // Convert macs
    for (int i = 0; i < request->macs_size(); ++i)
    {
        auto mac = request->macs(i);
        string_to_uint256(mac, macs[i]);
    }

    // Convert hsig
    success &= string_to_uint256(request->hsig(), hsig);

    // Convert rt
    int i = 0;
    for (auto &rt : request->rts())
    {
        success &= string_to_uint256(rt, rts[i++]);
    }

    // Convert reward
    reward = request->reward();

    // Convert fee
    fee = request->fee();

    // Convert address last byte
    address_last_byte = request->addresslastbyte() & ((1 << 8) - 1);

    return success;
}

int print_proof_inputs(const std::array<libzcash::JSInput, ZC_NUM_JS_INPUTS> &inputs,
                       std::array<libzcash::SproutNote, ZC_NUM_JS_OUTPUTS> &out_notes,
                       uint64_t vpub_old,
                       uint64_t vpub_new,
                       const RtArray &rts,
                       uint256 &h_sig,
                       uint252 &phi,
                       unsigned char &address_last_byte,
                       bool computeProof)
{
    cout << "Printing Proof's input" << '\n';
    for (auto &input : inputs)
    {
        cout << "input.key: " << input.key.inner().GetHex() << '\n';
        cout << "input.value: " << input.note.value() << '\n';
        cout << "input.note.a_pk: " << input.note.a_pk.GetHex() << '\n';
        cout << "input.note.r: " << input.note.r.GetHex() << '\n';
        cout << "input.note.rho: " << input.note.rho.GetHex() << '\n';
        cout << "input.note.cm: " << input.note.cm.GetHex() << '\n';
        cout << "input.note.nf: " << input.note.nf.GetHex() << '\n';
    }
    for (auto &output : out_notes)
    {
        cout << "output.value: " << output.value() << '\n';
        cout << "output.a_pk: " << output.a_pk.GetHex() << '\n';
        cout << "output.r: " << output.r.GetHex() << '\n';
        cout << "output.rho: " << output.rho.GetHex() << '\n';
        cout << "output.cm: " << output.cm.GetHex() << '\n';
        cout << "output.nf: " << output.nf.GetHex() << '\n';
    }
    for (auto &rt : rts)
    {
        cout << "rt: " << rt.GetHex() << '\n';
    }

    cout << "vpub_old: " << vpub_old << '\n';
    cout << "vpub_new: " << vpub_new << '\n';
    cout << "h_sig: " << h_sig.GetHex() << '\n';
    cout << "phi: " << phi.inner().GetHex() << '\n';
    cout << "address_last_byte: " << int(address_last_byte) << '\n';
    cout << "computeProof: " << computeProof << '\n';
    return 0;
}

class ZksnarkImpl final : public Zksnark::Service
{
    Status Prove(ServerContext *context, const ProveRequest *request, ProveReply *reply) override
    {
        cout << "\n\n[Starting Prove], request->inputs_size(): " << request->inputs_size() << "\n";
        ProveInputs inputs;
        ProveOutnotes out_notes;
        uint256 hsig;
        uint252 phi;
        RtArray rts;
        uint64_t reward, fee;
        unsigned char address_last_byte = 0;
        bool success = transform_prove_request(request, inputs, out_notes, hsig, phi, rts, reward, fee, address_last_byte);
        cout << "transform_prove_request status: " << success << '\n';

        if (!success) {
            reply->set_success(false);
            return Status::OK;
        }

        bool compute_proof = true;
        uint64_t vpub_old = reward;
        uint64_t vpub_new = fee;
        print_proof_inputs(inputs, out_notes, vpub_old, vpub_new, rts, hsig, phi, address_last_byte, compute_proof);
        // production make real proof
//        libzcash::PHGRProof proof;
//        success = js->prove(inputs, out_notes, vpub_old, vpub_new, rts, hsig, phi, address_last_byte, compute_proof, proof);
        // testing make default proof
        libzcash::PHGRProof proof;
        sleep(1);
        // print_proof(proof);

        zksnark::PHGRProof *zk_proof = new zksnark::PHGRProof();
        success &= transform_prove_reply(proof, *zk_proof);
        cout << "transform_prove_reply status: " << success << '\n';
        cout << "setting allocated_proof\n";
        reply->set_success(success);
        reply->set_allocated_proof(zk_proof);
        return Status::OK;
    }

    Status Verify(ServerContext *context, const VerifyRequest *request, VerifyReply *reply) override
    {
        cout << "\n\n[Starting Verify]\n\n";
        libzcash::PHGRProof proof;
        uint256 hsig;
        RtArray rts;
        NullifierArray nullifiers;
        CommitmentArray commitments;
        NullifierArray macs;
        uint64_t reward, fee;
        unsigned char address_last_byte = 0;
        bool success = transform_verify_request(request, proof, nullifiers, commitments, macs, hsig, rts, reward, fee, address_last_byte);
        cout << "transform_verify_request status: " << success << '\n';

        uint64_t vpub_old = reward;
        uint64_t vpub_new = fee;
        bool valid = false;

        // testing not verify proof
        // if (success) {
        //    success = js->verify(proof, macs, nullifiers, commitments, vpub_old, vpub_new, rts, hsig, address_last_byte, valid);
        // }

        // testing not verify proof
        success = true;
        valid = true;
        //

        reply->set_success(success);
        reply->set_valid(valid);
        return Status::OK;
    }
};

void RunServer(string &verifying_key, string &proving_key)
{
    // Creating zksnark circuit and load params
    cout << "Key: " << verifying_key << " " << proving_key << '\n';
    js = ZCJoinSplit::Prepared(verifying_key, proving_key);
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

int test_merkle_tree()
{
    uint256 x, y;
    bool ok = string_to_uint256("d26356e6f726dfb4c0a395f3af134851139ce1c64cfed3becc3530c8c8ad5660", x);
    ok &= string_to_uint256("5aaf71f995db014006d630dedf7ffcbfa8854055e6a8cc9ef153629e3045b7e1", y);
    if (!ok)
    {
        cout << "Fail to parse hash\n";
        return 0;
    }
    libzcash::SHA256Compress h1(x), h2(y);

    ZCIncrementalMerkleTree tree;
    // tree.append(h1);
    // tree.append(h2);
    std::deque<libzcash::SHA256Compress> filler_hash;
    auto rt = tree.root();
    cout << rt.GetHex() << '\n';
    return 0;
}

int generate_params() {
    ZCJoinSplit::Generate("/tmp/r1cs.params", "/tmp/verifying-lastbyte.key", "/tmp/proving-lastbyte.key");
    return 0;
}

int main(int argc, char const *argv[])
{
    // generate_params();
    // return 0;

    string verifying_key = "./verifying.key";
    string proving_key = "./proving.key";
    if (argc > 2) {
        verifying_key = string(argv[1]);
        proving_key = string(argv[2]);
    }
    RunServer(verifying_key, proving_key);
    return 0;
}
