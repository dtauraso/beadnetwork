













export interface PickOptions {
  excludeId?: string;
  nodesOnly?: boolean;
  ringOnly?: boolean;
  handholdOnly?: boolean;

  edgeOnly?: boolean;
}


export type PickFn = (ndcX: number, ndcY: number, opts?: PickOptions) => string | null;


export type PickRef = React.MutableRefObject<PickFn | null>;
