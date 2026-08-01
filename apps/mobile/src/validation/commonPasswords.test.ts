import { COMMON_PASSWORD_DENYLIST } from "./commonPasswords";

declare const __dirname: string;
declare function require(moduleName: string): unknown;

const { readFileSync } = require("node:fs") as {
  readFileSync(path: string, encoding: "utf8"): string;
};
const { resolve } = require("node:path") as {
  resolve(...paths: string[]): string;
};

describe("common-password parity", () => {
  it("matches the backend embedded corpus exactly", () => {
    const backendCorpus = readFileSync(
      resolve(
        __dirname,
        "../../../api/internal/security/common_passwords.txt",
      ),
      "utf8",
    )
      .split("\n")
      .map((value) => value.trim().normalize("NFC").toLowerCase())
      .filter((value) => value !== "" && !value.startsWith("#"));
    expect([...COMMON_PASSWORD_DENYLIST].sort()).toEqual(backendCorpus.sort());
  });
});
