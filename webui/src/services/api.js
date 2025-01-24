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
        return apiCall(`/conversations/${conversationId}/messages`)
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

    getConversation: async (conversationId) => {
        return await apiCall(`/conversations/${conversationId}`)
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
        return await apiCall(`/conversations/${conversationId}/messages/${messageId}/reactions`, {
            method: 'POST',
            body: JSON.stringify({ reaction: emoji })
        })
    }
}