import { BaseAPI, route } from "../base";
import type { EntitySummary } from "../types/data-contracts";
import type { PaginationResult } from "../types/non-generated";

export class AssetsApi extends BaseAPI {
  async get(id: string, page = 1, pageSize = 50) {
    return await this.http.get<PaginationResult<EntitySummary>>({
      url: route(`/assets/${id}`, { page, pageSize }),
    });
  }

  getMovementHistory(assetId: string, limit = 10) {
    return this.http.get<AssetMovementHistory[]>({
      url: route(`/assets/${assetId}/movement-history`, { limit }),
    });
  }

  createMovement(assetId: string, body: AssetMovementCreate) {
    return this.http.post<AssetMovementCreate, AssetMovementHistory>({
      url: route(`/assets/${assetId}/movement-history`),
      body,
    });
  }
}

export interface AssetMovementCreate {
  toLocationId?: string;
  movementType: "Transfer" | "Check Out" | "Check In" | "Return" | "Maintenance";
  movedByUserId?: string;
  otherPersonName?: string;
  otherPersonEmail?: string;
  otherPersonDepartment?: string;
  notes?: string;
  previousStatus?: string;
  newStatus?: string;
  checkedOutDueDate?: string;
  returnedDateTime?: string;
  approvedByUserId?: string;
  conditionBeforeMove?: string;
  conditionAfterMove?: string;
  overrideCheckedOutLock?: boolean;
}

export interface AssetMovementHistory {
  movementHistoryId: string;
  assetId: number;
  fromLocationId?: string;
  toLocationId?: string;
  movementType: string;
  movedByUserId?: string;
  otherPersonName: string;
  otherPersonEmail: string;
  otherPersonDepartment: string;
  movementDateTime: string;
  notes: string;
  createdByUserId: string;
  previousStatus: string;
  newStatus: string;
}
