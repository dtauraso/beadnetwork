function startRewritten(before, after) {
  return AXES.some(k => before.start[k] !== after.start[k]);
}

function sameButtons(run) {
  return run.sequence + " " + run.presses.map(b => b.k + b.sign).sort().join(" ");
}

function orderCheck(runs) {
  const groups = {};
  for (const run of runs) (groups[sameButtons(run)] = groups[sameButtons(run)] || []).push(run);
  return Object.values(groups).filter(list => list.length > 1).map(list => {
    const links = list[0].before.links.map((link, i) => {
      const deltas = list.map(run => run.after.links[i].delta);
      const rewritten = list.filter(run => startRewritten(run.before.links[i], run.after.links[i]));
      return {
        name: link.from + "→" + link.to,
        deltas,
        sameDelta: deltas.every(d => sameTicks(d, deltas[0])),
        rewritten: rewritten.length,
        deltaKept: rewritten.every(run => sameTicks(run.before.links[i].delta, run.after.links[i].delta)),
      };
    });
    return { sequence: list[0].sequence, orders: list.map(run => run.presses), links, ok: links.every(l => l.sameDelta && l.deltaKept) };
  });
}
