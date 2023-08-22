// @generated by protoc-gen-connect-es v0.13.0 with parameter "target=js+dts"
// @generated from file objectives/v1alpha1/objectives.proto (package objectives.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { GetAlertsRequest, GetAlertsResponse, GetStatusRequest, GetStatusResponse, GraphDurationRequest, GraphDurationResponse, GraphErrorBudgetRequest, GraphErrorBudgetResponse, GraphErrorsRequest, GraphErrorsResponse, GraphRateRequest, GraphRateResponse, ListRequest, ListResponse } from "./objectives_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service objectives.v1alpha1.ObjectiveService
 */
export const ObjectiveService = {
  typeName: "objectives.v1alpha1.ObjectiveService",
  methods: {
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveService.List
     */
    list: {
      name: "List",
      I: ListRequest,
      O: ListResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveService.GetStatus
     */
    getStatus: {
      name: "GetStatus",
      I: GetStatusRequest,
      O: GetStatusResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveService.GetAlerts
     */
    getAlerts: {
      name: "GetAlerts",
      I: GetAlertsRequest,
      O: GetAlertsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveService.GraphErrorBudget
     */
    graphErrorBudget: {
      name: "GraphErrorBudget",
      I: GraphErrorBudgetRequest,
      O: GraphErrorBudgetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveService.GraphRate
     */
    graphRate: {
      name: "GraphRate",
      I: GraphRateRequest,
      O: GraphRateResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveService.GraphErrors
     */
    graphErrors: {
      name: "GraphErrors",
      I: GraphErrorsRequest,
      O: GraphErrorsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveService.GraphDuration
     */
    graphDuration: {
      name: "GraphDuration",
      I: GraphDurationRequest,
      O: GraphDurationResponse,
      kind: MethodKind.Unary,
    },
  }
};

/**
 * @generated from service objectives.v1alpha1.ObjectiveBackendService
 */
export const ObjectiveBackendService = {
  typeName: "objectives.v1alpha1.ObjectiveBackendService",
  methods: {
    /**
     * @generated from rpc objectives.v1alpha1.ObjectiveBackendService.List
     */
    list: {
      name: "List",
      I: ListRequest,
      O: ListResponse,
      kind: MethodKind.Unary,
    },
  }
};

