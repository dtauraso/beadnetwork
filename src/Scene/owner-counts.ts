import { makeLeafValues } from "../valuefile/leaf-values";
import {
  OWNER_COUNTS_VALUE_NAMES, type OwnerCountsValueName,
} from "./owner-counts-values-gen";

const values = makeLeafValues<OwnerCountsValueName>(
  "Scene/owner-counts-paths",
  OWNER_COUNTS_VALUE_NAMES,
);

export function ownerCounts(): { nodes: number; edges: number } {
  return {
    nodes: values.i32("nodes"),
    edges: values.i32("edges"),
  };
}
