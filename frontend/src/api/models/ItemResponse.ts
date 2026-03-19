/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Todo item
 */
export type ItemResponse = {
    /**
     * Unique identifier for the given item.
     */
    id: string;
    name?: string;
    /**
     * Position for sort
     */
    position?: number;
    /**
     * The date that the item was created.
     */
    created_at?: string;
    /**
     * Done flag
     */
    done?: boolean;
};

