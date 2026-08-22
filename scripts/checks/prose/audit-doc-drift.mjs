#!/usr/bin/env node

import { readFileSync, existsSync } from "node:fs";
import { execSync } from "node:child_process";
import { join, dirname } from "node:path";

const root = execSync("git rev-parse --show-toplevel").toString().trim();

const mdFiles = execSync(
  `git ls-files '*.md' ':!:**/node_modules/**' ':!:docs/planning/**'`,
  { cwd: root },
).toString().trim().split("\n");

const srcFiles = execSync(
  `git ls-files '*.go' '*.ts' '*.tsx' '*.sh' '*.mjs' '*.yml' '*.yaml' '.gitignore' 'Makefile' '**/.gitignore' '**/Makefile' ':!:**/node_modules/**' ':!:docs/planning/**' ':!:**/*_generated.go' ':!:**/out/**'`,
  { cwd: root },
).toString().trim().split("\n");

const linkRe = /\]\(([^)#?]+)(?:[#?][^)]*)?\)/g;

const inlineRe = /`([^`\s]*\/[^`\s]+\.(?:md|ts|tsx|go|json|svg|sh|mjs|yml|yaml))`/g;

const srcRefRe =
  /\b(docs\/[A-Za-z0-9_./-]+\.(?:md|html)|MODEL\.md|CLAUDE\.md)\b/g;

const HISTORY_RE = /(gone|removed|retired|deleted|erased|obsolete|legacy|superseded|replaced|no longer|used to|formerly|was |were |since deleted|do not re-|don.t re-|never existed|no such|there is no)/i;

let fail = 0;

function lineOfFactory(text) {
  const lines = text.split("\n");
  return (idx) => {
    let acc = 0;
    for (let i = 0; i < lines.length; i++) {
      acc += lines[i].length + 1;
      if (idx < acc) return lines[i];
    }
    return "";
  };
}

function checkRef(file, abs, ref, idx, lineOf) {
  if (/^https?:\/\//.test(ref) || ref.startsWith("mailto:")) return;
  if (/[*<>{}]/.test(ref)) return; 
  if (HISTORY_RE.test(lineOf(idx))) return; 

  const rootCandidate = join(root, ref);
  const relCandidate = join(dirname(abs), ref);
  if (!existsSync(rootCandidate) && !existsSync(relCandidate)) {
    console.log(`doc-drift: ${file}: broken reference '${ref}'`);
    fail = 1;
  }
}

for (const file of mdFiles) {
  const abs = join(root, file);
  if (!existsSync(abs)) continue; 
  const text = readFileSync(abs, "utf8");
  const lineOf = lineOfFactory(text);
  const checks = [...text.matchAll(linkRe), ...text.matchAll(inlineRe)];
  for (const m of checks) checkRef(file, abs, m[1], m.index, lineOf);
}

for (const file of srcFiles) {
  if (!file) continue;
  const abs = join(root, file);

  if (!existsSync(abs)) continue;
  const text = readFileSync(abs, "utf8");
  const lineOf = lineOfFactory(text);
  for (const m of text.matchAll(srcRefRe)) {
    checkRef(file, abs, m[1], m.index, lineOf);
  }
}

process.exit(fail);
