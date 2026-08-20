import {
  RECV_STRIDE, FIRE_STRIDE, SEND_STRIDE, ARRIVE_STRIDE, BREADCRUMB_STRIDE,
} from "../../Buffer/buffer-layout";

export const STR_DECODER = new TextDecoder();

export interface EventSection {
  count: number;
  view: DataView;
}

export interface DecodedEvents {
  recv: EventSection;
  fire: EventSection;
  send: EventSection;
  arrive: EventSection;
  breadcrumb: EventSection;

  breadcrumbTextView: DataView;
}

function emptySection(buf: ArrayBuffer): EventSection {
  return { count: 0, view: new DataView(buf, buf.byteLength, 0) };
}

function emptyEvents(buf: ArrayBuffer): DecodedEvents {
  const e = emptySection(buf);
  return { recv: e, fire: e, send: e, arrive: e, breadcrumb: e, breadcrumbTextView: e.view };
}

export function decodeTrailingEvents(buf: ArrayBuffer, offset: number): DecodedEvents {
  const strides = [RECV_STRIDE, FIRE_STRIDE, SEND_STRIDE, ARRIVE_STRIDE, BREADCRUMB_STRIDE];
  const sections: EventSection[] = [];
  let off = offset;
  for (const stride of strides) {
    if (buf.byteLength < off + 4) return emptyEvents(buf);
    const count = new DataView(buf, off, 4).getUint32(0, true);
    const bytes = count * stride;
    if (buf.byteLength < off + 4 + bytes) return emptyEvents(buf);
    sections.push({ count, view: new DataView(buf, off + 4, bytes) });
    off += 4 + bytes;
  }
  return {
    recv: sections[0]!, fire: sections[1]!, send: sections[2]!,
    arrive: sections[3]!, breadcrumb: sections[4]!,
    breadcrumbTextView: new DataView(buf, off, buf.byteLength - off),
  };
}
