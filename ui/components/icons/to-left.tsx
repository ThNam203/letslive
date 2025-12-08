import React from 'react';
import { BaseIcon } from './base-icon';
import { IconProp } from '@/types/icon-prop';

function IconToLeft(props: IconProp) {
    return (
        <BaseIcon 
            width="1.5rem" 
            height="1.5rem" 
            viewBox="0 0 48 48"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="4"
            {...props}
        >
            <g fill="none">
                <path d="M14 23.9917H42"/>
                <path d="M26 36L14 24L26 12"/>
                <path d="M5 36V12"/>
            </g>
        </BaseIcon>
    );
}

export default IconToLeft;