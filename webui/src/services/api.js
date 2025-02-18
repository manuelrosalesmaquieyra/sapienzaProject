const API_URL = import.meta.env.VITE_API_BASE_URL || `${window.location.protocol}//${window.location.hostname}:3000`

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

    // Send message
    async sendMessage(conversationId, content) {
        return apiCall(`/conversations/${conversationId}/messages`, {
            method: 'POST',
            body: JSON.stringify({ content })
        })
    },

    // Add this function
    async createConversation(participantUsername) {
        console.log('Creating conversation with:', participantUsername);
        return apiCall('/conversations', {
            method: 'POST',
            body: JSON.stringify({
                participants: [participantUsername]
            })
        });
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
        console.log('Forwarding message:', {
            conversationId,
            messageId,
            url: `/conversations/${conversationId}/messages/${messageId}/forward`
        });
        return apiCall(`/conversations/${conversationId}/messages/${messageId}/forward`, {
            method: 'POST'
        });
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
            if (error.includes('same as current')) {
                throw new Error('This is already your current username');
            } else if (error.includes('already taken')) {
                throw new Error('This username is already taken, please choose a different one');
            }
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
    },
    
    // Create a new group
    createGroup: async (name, members) => {
        return apiCall('/groups', {
            method: 'POST',
            body: JSON.stringify({
                name: name,
                members: members  // Array of usernames
            })
        })
    },

    // Update group name
    updateGroupName: async (groupId, newName) => {
        return apiCall(`/groups/${groupId}`, {
            method: 'POST',
            body: JSON.stringify({
                new_name: newName  // Note: it's new_name, not name
            })
        })
    },

    // Update group photo
    updateGroupPhoto: async (groupId, file) => {
        const formData = new FormData()
        formData.append('photo', file)
        
        const response = await fetch(`${API_URL}/groups/${groupId}/photo`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`
            },
            body: formData
        })

        if (!response.ok) {
            const text = await response.text()
            throw new Error(text)
        }

        return response.json()
    },

    // Leave group
    leaveGroup: async (groupId) => {
        return apiCall(`/groups/${groupId}/leave`, {
            method: 'POST'
        })
    },

    // Add these methods to your api object
    replyToMessage: async (conversationId, messageId, content) => {
        return apiCall(`/conversations/${conversationId}/messages/${messageId}/reply`, {
            method: 'POST',
            body: JSON.stringify({ content })
        });
    },

    sendImageMessage: async (conversationId, imageFile) => {
        const formData = new FormData();
        formData.append('image', imageFile);
        
        const response = await fetch(`${API_URL}/conversations/${conversationId}/image-message`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`
            },
            body: formData
        });

        if (!response.ok) {
            const text = await response.text();
            throw new Error(text);
        }

        return response.json();
    },

    // Add this new method to the api object
    uploadProfilePhoto: async (username, file) => {
        const formData = new FormData()
        formData.append('photo', file)
        
        const response = await fetch(`${API_URL}/users/${username}/photo`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('sessionId')}`
            },
            body: formData
        })

        if (!response.ok) {
            const text = await response.text()
            throw new Error(text)
        }

        return response.json()
    },

    // Fix the checkUserExists method
    async checkUserExists(username) {
        try {
            const response = await fetch(`${API_URL}/users/${username}/exists`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('sessionId')}`,
                    'Content-Type': 'application/json'
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                return data.exists;
            }
            return false;
        } catch (err) {
            console.error('Error checking user:', err);
            return false;
        }
    },

    // Update the getAllUsers method
    getAllUsers: async () => {
        return apiCall('/allusers', {
            method: 'GET'
        })
    },
}