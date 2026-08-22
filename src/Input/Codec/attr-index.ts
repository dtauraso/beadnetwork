import { IN_UPDATE_ATTRS, IN_UPDATE_KINDS, IN_KIND_EDIT_UPDATE } from "./input-layout-gen";
import { ByteWriter, enumIndex } from "./byte-writer";

export function attrIndex(attr: string): number {
  const i = (IN_UPDATE_ATTRS as readonly string[]).indexOf(attr);
  if (i < 0) {
    throw new Error(
      `attrIndex: "${attr}" is not in the updateAttrs list of INPUT_LAYOUT_FINGERPRINT, so ` +
        `nothing on the wire can carry it; add it to the fingerprint in the same commit`,
    );
  }
  return i;
}

export function editUpdate(entity: string, attr: string): ByteWriter {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, entity));
  w.u8(attrIndex(attr));
  return w;
}
