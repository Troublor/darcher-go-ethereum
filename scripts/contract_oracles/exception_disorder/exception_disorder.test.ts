import {expect} from "chai";
import * as path from "path";
import * as shell from "shelljs";
import {ContractVulType} from "@darcher/rpc";

describe("exception disorder oracle", () => {
    it('should pass test case 1', async function () {
        let testCaseDir = path.join(__dirname, "1");
        let binDir = path.join(__dirname, "..", "..", "..", "build", "bin");
        let resultFile = path.join(testCaseDir, "result.json");
        let seg = [`${binDir}/evm t8n`];
        seg.push(`--input.alloc=${path.join(testCaseDir, "alloc.json")}`);
        seg.push(`--input.txs=${path.join(testCaseDir, "txs.json")}`);
        seg.push(`--input.env=${path.join(testCaseDir, "env.json")}`);
        seg.push(`--output.alloc=${path.join(testCaseDir, "out-alloc.json")}`);
        seg.push(`--output.result=${path.join(testCaseDir, "out-result.json")}`);
        seg.push(`--state.chainid 1337`);
        seg.push(`--analyze --analyze.report=${resultFile}`);
        seg.push(`--verbosity 2`)
        let cmd = seg.join(" ");
        shell.exec(`${cmd}`)
        let results = await import(resultFile);
        expect(results).to.be.lengthOf(2);
        expect(results[0].type).to.be.equal(ContractVulType.GASLESS_SEND);
        expect(results[0].tx_hash).to.be.equal("0xfdf97294b132e59f1c4385004d9a9fa2eab857c04ce440e8db701b8fbae41323");
        expect(results[1].type).to.be.equal(ContractVulType.EXCEPTION_DISORDER);
        expect(results[1].tx_hash).to.be.equal("0xfdf97294b132e59f1c4385004d9a9fa2eab857c04ce440e8db701b8fbae41323");
    });
})