// Contract: the probe-trace setting key read by probe-files.ts (via
// vscode.workspace.getConfiguration(SECTION).get(KEY, false)) must actually exist in
// package.json's contributes.configuration.properties. getConfiguration().get(key, default)
// returns the DEFAULT SILENTLY if the joined key doesn't match a declared property — a
// typo or rename on either side leaves the gate permanently stuck off with no error
// anywhere. Also assert default=false/type=boolean, since default-true would silently
// reintroduce the gigabyte trace logs.

import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import {
  PROBE_TRACE_SETTING_KEY,
  PROBE_TRACE_SETTING_SECTION,
} from "../../src/probe-files";

const PACKAGE_JSON_PATH = join(__dirname, "../../package.json");

describe("probe-trace-setting parity contract", () => {
  it("PROBE_TRACE_SETTING_SECTION.PROBE_TRACE_SETTING_KEY is declared in package.json", () => {
    const pkg = JSON.parse(readFileSync(PACKAGE_JSON_PATH, "utf8"));
    const properties = pkg.contributes.configuration.properties;
    const joinedKey = `${PROBE_TRACE_SETTING_SECTION}.${PROBE_TRACE_SETTING_KEY}`;

    expect(properties).toHaveProperty(joinedKey);
    expect(properties[joinedKey].default).toBe(false);
    expect(properties[joinedKey].type).toBe("boolean");
  });
});
