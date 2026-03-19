/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ItemPatch } from '../models/ItemPatch';
import type { ItemPost } from '../models/ItemPost';
import type { ItemResponse } from '../models/ItemResponse';
import type { ItemsResponse } from '../models/ItemsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ItemsService {
    /**
     * Get Item Info by Item ID
     * Retrieve the information of the item with the matching item ID.
     * @param itemId Item ID
     * @returns ItemResponse OK
     * @throws ApiError
     */
    public static getItemsItemId(
        itemId: string,
    ): CancelablePromise<ItemResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/items/{item_id}',
            path: {
                'item_id': itemId,
            },
            errors: {
                401: `Unauthorized`,
                403: `Forbidden`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Update Item
     * Update the information of an existing item.
     * @param itemId Item ID
     * @param requestBody Patch user properties to update.
     * @returns ItemResponse User Updated
     * @throws ApiError
     */
    public static patchItemsItemid(
        itemId: string,
        requestBody?: ItemPatch,
    ): CancelablePromise<ItemResponse> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/items/{item_id}',
            path: {
                'item_id': itemId,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                401: `Unauthorized`,
                403: `Forbidden`,
                404: `User Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Create New User
     * Create a new item.
     * @param requestBody Post the necessary fields for the API to create a new item.
     * @returns ItemResponse Item Created
     * @throws ApiError
     */
    public static postItems(
        requestBody?: ItemPost,
    ): CancelablePromise<ItemResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/items',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Missing Required Information`,
                401: `Unauthorized`,
                403: `Forbidden`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Get all items
     * Get all items
     * @param perPage Count items per page
     * @param page Page number
     * @param sort Sort by fields
     * @param fields Retrieve certain fields
     * @param done Filter by done
     * @param requestBody
     * @returns ItemsResponse OK
     * @throws ApiError
     */
    public static getItems(
        perPage: number = 20,
        page: number = 1,
        sort?: string,
        fields?: string,
        done?: boolean,
        q?: string,
        requestBody?: any,
    ): CancelablePromise<ItemsResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/items',
            query: {
                '_per_page': perPage,
                '_page': page,
                '_sort': sort,
                '_fields': fields,
                'done': done,
                'q': q,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                401: `Unauthorized`,
                403: `Forbidden`,
                500: `Internal Server Error`,
            },
        });
    }
}
