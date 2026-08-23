
export interface NodeDef {
  bg: string;
  border: string;
  text: string;
  minWidth?: number;
  shape?: string;
  fill?: string;
  stroke?: string;
  width?: number;
  height?: number;
  desc?: string;
  inputs?: { name: string; kind: string; isMulti?: boolean }[];
  outputs?: { name: string; kind: string; isMulti?: boolean }[];
}
