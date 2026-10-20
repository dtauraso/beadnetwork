const T_WHY_EDITS = [
  'Loading the example places a, b and c where they belong, with its two trees, a→b and b→c.',
  'Changing a start moves b; c stays where it is, and b→c\'s start is re-measured from b\'s new place.',
  'The same as φ, for θ.',
  'Changing the length moves b outward.',
  'Another φ edit moves b again.',
  'Another θ edit moves b again.',
  'A new node and tree get placed. A single a→b edit then moves b only.',
  'A θ edit with four nodes moves b only.',
  'A φ edit with four nodes moves b only.',
  'Going back to φ 3 puts b where φ 3 with θ 4 puts it.',
  'A start that is not a whole number is taken as it is: b moves, and b→c\'s start is re-measured to where c still is.',
  'An edit further down the line, on a tree other than a→b, moves c without touching a, b or d.',
  'The state is packed (bPack), reloaded and rewired (bWire), and comes back exactly the same: positions and trees.',
  'The reloaded trees still work: their links are rebuilt well enough for an edit to move things as the original did.',
  'Dropping a tree takes it out of both nodes\' lists.',
  'The same for a→b.',
  'Dropping a node drops every tree into and out of it, and the remaining nodes stay as they were.',
];

const T_WHY_SAVES = [
  'The directions the builder works out for each row start right.',
  'The directions after one edit.',
  'A save in the old format, from before the rows moved into the trees, still loads; its a→c copy loads as an angle edge measured from where a and c are, so c does not move.',
  'An old save, once loaded, still edits correctly.',
  'Repacking an old save in the new format loses nothing, and the directions still match the step before.',
  'A save that went from the old format to the new one and back in still edits correctly.',
];
