const API_URL = 'http://localhost:3000'

async function apiCall(endpoint, options = {}) {
    const sessionId = localStorage.getItem('sessionId')
    
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${sessionId}`
    }

    // Log the exact request details
    console.log('Making request:', {
        url: `${API_URL}${endpoint}`,
        method: options.method || 'GET',
        headers,
        sessionId
    })

    const response = await fetch(`${API_URL}${endpoint}`, {
        ...options,
        mode: 'cors',
        headers: {
            ...headers,
            ...options.headers
        }
    })

    if (!response.ok) {
        console.error('Response error:', {
            status: response.status,
            statusText: response.statusText
        })
        const text = await response.text()
        throw new Error(`API call failed: ${text}`)
    }

    // Handle empty response for DELETE
    if (response.status === 204) {
        return null
    }

    return response.json()
}

export const api = {
    // Get user's conversations
    async getConversations(username) {
        return apiCall(`/users/${username}/conversations`)
    },

    // Get conversation messages
    async getConversationMessages(conversationId) {
        const response = await fetch(`${API_URL}/conversations/${conversationId}/messages`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`
            }
        })
        
        if (!response.ok) throw new Error('Failed to fetch messages')
        const data = await response.json()
        return data
    },

    // Create group
    async createGroup(name, members) {
        return apiCall('/groups', {
            method: 'POST',
            body: JSON.stringify({ name, members })
        })
    },

    // Send message
    async sendMessage(conversationId, content) {
        return apiCall(`/conversations/${conversationId}/messages`, {
            method: 'POST',
            body: JSON.stringify({ content })
        })
    },

    // Add this function
    async createConversation(participantUsername) {
        const requestBody = {
            participants: [participantUsername]
        };
        console.log('Sending conversation request:', requestBody);
        
        const result = await apiCall('/conversations', {
            method: 'POST',
            body: JSON.stringify(requestBody)
        });
        console.log('Create conversation response:', result);
        return result;
    },

    getConversationDetails: async (conversationId) => {
        return apiCall(`/conversations/${conversationId}/details`)
    },

    getConversation: async (conversationId) => {
        return apiCall(`/conversations/${conversationId}`)
    },

    deleteMessage: async (conversationId, messageId) => {
        return await apiCall(`/conversations/${conversationId}/messages/${messageId}`, {
            method: 'DELETE'
        })
    },
    
    forwardMessage: async (conversationId, messageId) => {
        return await apiCall(`/conversations/${conversationId}/messages/${messageId}`, {
            method: 'POST'
        })
    },

    addReaction: async (messageId, emoji, conversationId) => {
        const response = await fetch(`${API_URL}/conversations/${conversationId}/messages/${messageId}/reactions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`
            },
            body: JSON.stringify({ 
                reaction: emoji 
            })
        })
        if (!response.ok) throw new Error('Failed to add reaction')
        
        // Don't try to parse JSON if there's no content
        const text = await response.text()
        return text.length > 0 ? JSON.parse(text) : {}
    },

    deleteReaction: async (messageId, conversationId) => {
        const response = await fetch(`${API_URL}/conversations/${conversationId}/messages/${messageId}/reactions`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`
            }
        })
        if (!response.ok) throw new Error('Failed to delete reaction')
        return {}
    },

    updateUsername: async (currentUsername, newUsername) => {
        const response = await fetch(`${API_URL}/users/${currentUsername}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`,
            },
            body: JSON.stringify({
                new_name: newUsername
            })
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || 'Failed to update username');
        }

        const data = await response.json();
        return data;
    },

    updateProfilePhoto: async (username, photoUrl) => {
        const payload = {
            photo_url: photoUrl === null ? "" : photoUrl  // Convert null to empty string
        }
        
        return apiCall(`/users/${username}/photo`, {
            method: 'POST',
            body: JSON.stringify(payload)
        })
    },

    async getUserProfile(username) {
        console.log('Fetching profile for:', username)
        const response = await fetch(`${API_URL}/users/${username}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`,
            }
        });

        if (!response.ok) {
            console.error('Profile fetch failed:', response.status)
            throw new Error('Failed to fetch user profile');
        }

        const data = await response.json()
        console.log('Profile data received:', data)
        return data
    }
}