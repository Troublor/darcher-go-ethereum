import * as shell from "shelljs";
import * as path from "path";
import {expect} from "chai";
import {ContractVulType} from "@darcher/rpc";

describe("reentrancy oracle", () => {
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
        expect(results).to.be.lengthOf(1);
        expect(results[0].type).to.be.equal(ContractVulType.REENTRANCY);
        expect(results[0].tx_hash).to.be.equal("0xdf93fdb7502a71fbe17eebdfdfd2aaf3b4aeb31cd194a30704419866ec295db7");
    });
});