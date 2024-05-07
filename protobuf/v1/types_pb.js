// @generated by protoc-gen-es v1.9.0
// @generated from file v1/types.proto (package proto.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3 } from "@bufbuild/protobuf";

/**
 * @generated from message proto.v1.HTTPRequest
 */
export const HTTPRequest = /*@__PURE__*/ proto3.makeMessageType(
  "proto.v1.HTTPRequest",
  () => [
    { no: 1, name: "body", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
    { no: 2, name: "method", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "endpoint_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "env", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
    { no: 6, name: "header", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: HeaderFields} },
    { no: 7, name: "runtime", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 8, name: "deployment_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 9, name: "id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message proto.v1.HeaderFields
 */
export const HeaderFields = /*@__PURE__*/ proto3.makeMessageType(
  "proto.v1.HeaderFields",
  () => [
    { no: 1, name: "fields", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ],
);

/**
 * @generated from message proto.v1.HTTPResponse
 */
export const HTTPResponse = /*@__PURE__*/ proto3.makeMessageType(
  "proto.v1.HTTPResponse",
  () => [
    { no: 1, name: "body", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
    { no: 2, name: "code", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 3, name: "request_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "header", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: HeaderFields} },
  ],
);

