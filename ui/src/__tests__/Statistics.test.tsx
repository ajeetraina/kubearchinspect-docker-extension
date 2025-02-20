import React from 'react';
import { render, screen } from '@testing-library/react';
import { Statistics } from '../components/Statistics';

describe('Statistics', () => {
    const mockResources = [
        { name: 'test1', kind: 'Deployment', namespace: 'default', isArmCompatible: true },
        { name: 'test2', kind: 'Deployment', namespace: 'default', isArmCompatible: false }
    ];

    it('renders statistics correctly', () => {
        render(<Statistics resources={mockResources} />);
        expect(screen.getByText('ARM Compatible')).toBeInTheDocument();
        expect(screen.getByText('1')).toBeInTheDocument();
    });
});