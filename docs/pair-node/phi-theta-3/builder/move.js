function bMove(builder, t) {
  if (t.ab.a.send(builder, t)) t.ab.b.moved(builder, [t]);
}

function bMoveAll(builder, trees) {
  trees.forEach(() => trees.forEach(t => t.ab.a.send(builder, t)));
  trees.forEach(t => t.ab.b.moved(builder, trees));
}
