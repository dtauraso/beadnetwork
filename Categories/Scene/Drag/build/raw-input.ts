export type RawPointerKind =
  | "pointerdown"
  | "pointermove"
  | "pointerup"
  | "wheel"
  | "home"
  | "delete"
  | "key";

export type RawHit = {
  kind: "port" | "handhold" | "node" | "edge" | "torus" | "empty";
  isInput: boolean;

  onRim: boolean;

  nodeRow: number;
  portRow: number;

  edgeRow: number;

  pointX: number;
  pointY: number;
  pointZ: number;
};

export type RawInputEvent = {
  kind: RawPointerKind;
  x: number;
  y: number;
  rectLeft: number;
  rectTop: number;
  rectWidth: number;
  rectHeight: number;
  button: number;
  ctrl: boolean;
  shift: boolean;
  alt: boolean;
  meta: boolean;
  deltaX: number;
  deltaY: number;
  hit: RawHit;
  key?: string;

  ballX: number;
  ballY: number;
  ballZ: number;

  ballPrevX: number;
  ballPrevY: number;
  ballPrevZ: number;
};
