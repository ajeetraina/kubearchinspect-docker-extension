import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import App from '../App';

describe('App', () => {
    it('renders without crashing', () => {
        render(<App />);
        expect(screen.getByText('Kubernetes ARM Inspector')).toBeInTheDocument();
    });

    it('handles inspect button click', async () => {
        render(<App />);
        const button = screen.getByText('Inspect Resources');
        fireEvent.click(button);
        // Add assertions for after click behavior
    });
});