import React from 'react';
import { BaseIcon } from './base-icon';
import { IconProp } from '@/types/icon-prop';

function IconCheck(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path fill="currentColor" d="M21 7L9 19l-5.5-5.5l1.41-1.41L9 16.17L19.59 5.59z"/>
        </BaseIcon>
    );
}

export default IconCheck;