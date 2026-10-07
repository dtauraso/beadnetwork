import { drawBox, canvasFont, roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
import { drawPill } from "../../Pills/pill";
import { cardBytes, cardF32, cardF32Run, cardI32Run, cardU8, cardU32Run, cardText } from "./panel-leaves";

const TITLE_FONT_PX = 12;
const TITLE_INK = "#222";

const CELL_FILL = "#222";
const CELL_EDIT_FILL = "#0d47a1";
const CELL_RADIUS = 3;
const CELL_PAD_X = 6;
const KEY_FONT_PX = 11;
const KEY_INK = "#e0e0e0";
const VAL_FONT_PX = 13;
const VAL_INK = "#fff";

export function cardDraftOpen(): boolean {
  const editing = cardBytes("fieldEditing");
  if (!editing) return false;
  for (let i = 0; i < editing.byteLength; i++) {
    if (editing.getUint8(i) !== 0) return true;
  }
  return false;
}

export function cardPanelKey(): string {
  const boxX = cardF32Run("boxX");
  const values = cardI32Run("fieldValue");
  const editing = cardBytes("fieldEditing");
  const draft = cardText("draftText");
  const editFlags: number[] = [];
  if (editing) for (let i = 0; i < editing.byteLength; i++) editFlags.push(editing.getUint8(i));
  return [
    cardF32("pillX"), cardF32("pillY"), cardF32("pillW"), cardU8("open"),
    boxX ? Array.from(boxX).join(".") : "",
    values ? Array.from(values).join(".") : "",
    editFlags.join(""),
    draft ? decodeAt(draft, 0, draft.length) : "",
  ].join(",");
}

function drawField(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
  key: string, value: string, editing: boolean,
): void {
  if (w <= 0 || h <= 0) return;
  roundRect(c, x, y, w, h, CELL_RADIUS);
  c.fillStyle = editing ? CELL_EDIT_FILL : CELL_FILL;
  c.fill();

  c.textBaseline = "middle";
  c.fillStyle = KEY_INK;
  c.font = canvasFont(KEY_FONT_PX, 600);
  c.textAlign = "left";
  c.fillText(key, x + CELL_PAD_X, y + h / 2);

  c.fillStyle = VAL_INK;
  c.font = canvasFont(VAL_FONT_PX, 700);
  c.textAlign = "right";
  c.fillText(value, x + w - CELL_PAD_X, y + h / 2);
}

export function drawCardPanel(c: CanvasRenderingContext2D): void {
  const pillText = cardText("pillText");
  const pw = cardF32("pillW");
  const ph = cardF32("pillH");
  if (pillText && pw > 0 && ph > 0) {
    drawPill(c, cardF32("pillX"), cardF32("pillY"), pw, ph, decodeAt(pillText, 0, pillText.length), cardU8("open") !== 0);
  }

  const boxX = cardF32Run("boxX");
  const boxY = cardF32Run("boxY");
  const boxW = cardF32Run("boxW");
  const boxH = cardF32Run("boxH");
  const headX = cardF32Run("headX");
  const headY = cardF32Run("headY");
  const headH = cardF32Run("headH");
  const titleText = cardText("titleText");
  const titleLen = cardU32Run("titleLen");
  if (!boxX || !boxY || !boxW || !boxH || !headX || !headY || !headH || !titleText || !titleLen) return;

  let titleOff = 0;
  for (let i = 0; i < boxX.length; i++) {
    drawBox(c, boxX[i]!, boxY[i]!, boxW[i]!, boxH[i]!);
    const title = decodeAt(titleText, titleOff, titleLen[i]!);
    titleOff += titleLen[i]!;
    c.fillStyle = TITLE_INK;
    c.font = canvasFont(TITLE_FONT_PX, 700);
    c.textAlign = "left";
    c.textBaseline = "middle";
    c.fillText(title, headX[i]!, headY[i]! + headH[i]! / 2);
  }

  const fieldX = cardF32Run("fieldX");
  const fieldY = cardF32Run("fieldY");
  const fieldW = cardF32Run("fieldW");
  const fieldH = cardF32Run("fieldH");
  const keyText = cardText("keyText");
  const keyLen = cardU32Run("keyLen");
  const values = cardI32Run("fieldValue");
  const editing = cardBytes("fieldEditing");
  const draftBytes = cardText("draftText");
  if (!fieldX || !fieldY || !fieldW || !fieldH || !keyText || !keyLen || !values || !editing) return;
  const draft = draftBytes ? decodeAt(draftBytes, 0, draftBytes.length) : "";

  let keyOff = 0;
  for (let i = 0; i < fieldX.length; i++) {
    const key = decodeAt(keyText, keyOff, keyLen[i]!);
    keyOff += keyLen[i]!;
    const isEditing = i < editing.byteLength && editing.getUint8(i) !== 0;
    const shown = isEditing ? `${draft}_` : String(values[i]!);
    drawField(c, fieldX[i]!, fieldY[i]!, fieldW[i]!, fieldH[i]!, key, shown, isEditing);
  }
}
