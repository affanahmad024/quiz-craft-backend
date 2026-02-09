# Frontend Integration Guide

This guide helps you integrate your frontend application with the QuizCraft backend API.

---

## 🚀 Quick Start

### 1. API Base URL

```javascript
const API_BASE_URL = 'http://localhost:8080/api/v1';
// Production: 'https://api.quizcraft.com/api/v1'
```

### 2. Install HTTP Client

```bash
npm install axios
# or
npm install fetch
```

### 3. Basic Setup

```javascript
// src/services/api.js
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle token refresh on 401
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Token expired, try refresh
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        try {
          const { data } = await axios.post(`${API_BASE_URL}/auth/refresh`, {
            refresh_token: refreshToken
          });
          localStorage.setItem('access_token', data.data.access_token);
          // Retry original request
          error.config.headers.Authorization = `Bearer ${data.data.access_token}`;
          return axios(error.config);
        } catch (refreshError) {
          // Refresh failed, logout
          localStorage.clear();
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

export default api;
```

---

## 🔐 Authentication Flow

### Signup & Login

```javascript
// src/services/authService.js
import api from './api';

export const authService = {
  // Signup
  async signup(email, name, password) {
    const { data } = await api.post('/auth/signup', {
      email,
      name,
      password
    });
    return data;
  },

  // Login
  async login(email, password) {
    const { data } = await api.post('/auth/login', {
      email,
      password
    });

    // Store tokens
    localStorage.setItem('access_token', data.data.access_token);
    localStorage.setItem('refresh_token', data.data.refresh_token);
    localStorage.setItem('user', JSON.stringify(data.data.user));

    return data.data.user;
  },

  // Logout
  async logout() {
    try {
      await api.post('/auth/logout');
    } finally {
      localStorage.clear();
      window.location.href = '/login';
    }
  },

  // Get current user
  getCurrentUser() {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  },

  // Check if authenticated
  isAuthenticated() {
    return !!localStorage.getItem('access_token');
  }
};
```

### React Authentication Hook

```javascript
// src/hooks/useAuth.js
import { createContext, useContext, useState, useEffect } from 'react';
import { authService } from '../services/authService';

const AuthContext = createContext();

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const user = authService.getCurrentUser();
    setUser(user);
    setLoading(false);
  }, []);

  const login = async (email, password) => {
    const user = await authService.login(email, password);
    setUser(user);
  };

  const logout = async () => {
    await authService.logout();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
```

---

## 📝 Quiz Management

```javascript
// src/services/quizService.js
import api from './api';

export const quizService = {
  // Create quiz
  async createQuiz(quizData) {
    const { data } = await api.post('/quizzes', quizData);
    return data.data;
  },

  // Get my quizzes
  async getMyQuizzes(page = 1, limit = 20) {
    const { data } = await api.get('/quizzes', {
      params: { page, limit }
    });
    return data.data;
  },

  // Get public quizzes
  async getPublicQuizzes(filters = {}) {
    const { data } = await api.get('/quizzes/public', {
      params: filters
    });
    return data.data;
  },

  // Get quiz with questions
  async getQuizFull(quizId) {
    const { data } = await api.get(`/quizzes/${quizId}/full`);
    return data.data;
  },

  // Update quiz
  async updateQuiz(quizId, updates) {
    const { data } = await api.patch(`/quizzes/${quizId}`, updates);
    return data.data;
  },

  // Publish quiz
  async publishQuiz(quizId) {
    const { data } = await api.post(`/quizzes/${quizId}/publish`);
    return data.data;
  },

  // Delete quiz
  async deleteQuiz(quizId) {
    await api.delete(`/quizzes/${quizId}`);
  },

  // Search quizzes
  async searchQuizzes(query, filters = {}) {
    const { data } = await api.get('/quizzes/search', {
      params: { q: query, ...filters }
    });
    return data.data;
  }
};
```

---

## ❓ Question Management

```javascript
// src/services/questionService.js
import api from './api';

export const questionService = {
  // Add question (JSON)
  async addQuestion(quizId, questionData) {
    const { data } = await api.post(
      `/quizzes/${quizId}/questions`,
      questionData
    );
    return data.data;
  },

  // Add question with image
  async addQuestionWithImage(quizId, questionData, imageFile) {
    const formData = new FormData();
    formData.append('image', imageFile);
    formData.append('question_data', JSON.stringify(questionData));

    const { data } = await api.post(
      `/quizzes/${quizId}/questions`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      }
    );
    return data.data;
  },

  // Add multiple questions
  async addQuestionsBulk(quizId, questions) {
    const { data } = await api.post(
      `/quizzes/${quizId}/questions/bulk`,
      { questions }
    );
    return data.data;
  },

  // Update question
  async updateQuestion(questionId, updates) {
    const { data } = await api.patch(
      `/quizzes/questions/${questionId}`,
      updates
    );
    return data.data;
  },

  // Delete question
  async deleteQuestion(questionId) {
    await api.delete(`/quizzes/questions/${questionId}`);
  }
};
```

---

## 🎯 Quiz Attempt Flow

```javascript
// src/services/attemptService.js
import api from './api';

export const attemptService = {
  // Start quiz attempt
  async startAttempt(quizId) {
    const { data } = await api.post(`/quizzes/${quizId}/attempts`);
    return data.data;
  },

  // Get attempt
  async getAttempt(attemptId) {
    const { data } = await api.get(`/attempts/${attemptId}`);
    return data.data;
  },

  // Submit answer
  async submitAnswer(attemptId, questionId, answerId) {
    const { data } = await api.post(`/attempts/${attemptId}/answers`, {
      question_id: questionId,
      answer_id: answerId
    });
    return data.data;
  },

  // Finish attempt
  async finishAttempt(attemptId) {
    const { data } = await api.post(`/attempts/${attemptId}/finish`);
    return data.data;
  },

  // Get results
  async getResults(attemptId) {
    const { data } = await api.get(`/attempts/${attemptId}/results`);
    return data.data;
  },

  // Get my attempts
  async getMyAttempts(filters = {}) {
    const { data } = await api.get('/attempts', { params: filters });
    return data.data;
  }
};
```

### React Quiz Component Example

```javascript
// src/components/QuizAttempt.jsx
import { useState, useEffect } from 'react';
import { attemptService } from '../services/attemptService';

export function QuizAttempt({ quizId }) {
  const [attempt, setAttempt] = useState(null);
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [selectedAnswer, setSelectedAnswer] = useState(null);

  useEffect(() => {
    startQuiz();
  }, [quizId]);

  const startQuiz = async () => {
    const attemptData = await attemptService.startAttempt(quizId);
    setAttempt(attemptData);
  };

  const submitAnswer = async () => {
    if (!selectedAnswer) return;

    await attemptService.submitAnswer(
      attempt.id,
      attempt.questions[currentQuestion].id,
      selectedAnswer
    );

    // Move to next question
    if (currentQuestion < attempt.questions.length - 1) {
      setCurrentQuestion(currentQuestion + 1);
      setSelectedAnswer(null);
    } else {
      finishQuiz();
    }
  };

  const finishQuiz = async () => {
    const results = await attemptService.finishAttempt(attempt.id);
    // Show results
    console.log('Quiz completed!', results);
  };

  if (!attempt) return <div>Loading...</div>;

  const question = attempt.questions[currentQuestion];

  return (
    <div>
      <h2>{question.question_text}</h2>
      {question.answers.map((answer) => (
        <div key={answer.id}>
          <input
            type="radio"
            name="answer"
            value={answer.id}
            checked={selectedAnswer === answer.id}
            onChange={(e) => setSelectedAnswer(e.target.value)}
          />
          <label>{answer.answer_text}</label>
        </div>
      ))}
      <button onClick={submitAnswer}>
        {currentQuestion < attempt.questions.length - 1 ? 'Next' : 'Finish'}
      </button>
    </div>
  );
}
```

---

## 🤖 AI Quiz Generation

```javascript
// src/services/aiService.js
import api from './api';

export const aiService = {
  // Generate from text
  async generateFromText(text, options) {
    const { data } = await api.post('/ai/generate-quiz-from-text', {
      text,
      options
    });
    return data.data;
  },

  // Generate from file
  async generateFromFile(file, options) {
    const formData = new FormData();
    formData.append('file', file);

    Object.keys(options).forEach((key) => {
      formData.append(`options[${key}]`, options[key]);
    });

    const { data } = await api.post('/ai/generate-quiz-from-file', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    return data.data;
  },

  // Extract from image/PDF
  async extractFromFile(file, extractImages = true) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('extract_images', extractImages);

    const { data } = await api.post('/ai/extract-quiz-from-file', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    return data.data;
  },

  // Check generation status
  async checkStatus(generationId) {
    const { data } = await api.get(`/ai/generation-status/${generationId}`);
    return data.data;
  },

  // Poll for completion
  async waitForCompletion(generationId) {
    return new Promise((resolve, reject) => {
      const interval = setInterval(async () => {
        try {
          const status = await this.checkStatus(generationId);

          if (status.status === 'completed') {
            clearInterval(interval);
            resolve(status);
          } else if (status.status === 'failed') {
            clearInterval(interval);
            reject(new Error('Generation failed'));
          }
        } catch (error) {
          clearInterval(interval);
          reject(error);
        }
      }, 3000); // Check every 3 seconds
    });
  }
};
```

---

## 🔔 Notifications (Real-time)

### REST API

```javascript
// src/services/notificationService.js
import api from './api';

export const notificationService = {
  // Get notifications
  async getNotifications(params = {}) {
    const { data } = await api.get('/notifications', { params });
    return data.data;
  },

  // Get unread count
  async getUnreadCount() {
    const { data } = await api.get('/notifications/unread-count');
    return data.data.count;
  },

  // Mark as read
  async markAsRead(notificationId) {
    await api.patch(`/notifications/${notificationId}/read`);
  },

  // Mark all as read
  async markAllAsRead() {
    await api.patch('/notifications/read-all');
  },

  // Delete notification
  async deleteNotification(notificationId) {
    await api.delete(`/notifications/${notificationId}`);
  }
};
```

### WebSocket Connection

```javascript
// src/services/notificationWebSocket.js
class NotificationWebSocket {
  constructor() {
    this.ws = null;
    this.listeners = [];
  }

  connect(token) {
    const wsUrl = `ws://localhost:8080/api/v1/notifications/ws?token=${token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('Connected to notifications');
    };

    this.ws.onmessage = (event) => {
      const notification = JSON.parse(event.data);
      this.listeners.forEach((listener) => listener(notification));
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('Disconnected from notifications');
      // Reconnect after 5 seconds
      setTimeout(() => this.connect(token), 5000);
    };
  }

  onNotification(callback) {
    this.listeners.push(callback);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

export const notificationWS = new NotificationWebSocket();
```

### React Hook

```javascript
// src/hooks/useNotifications.js
import { useState, useEffect } from 'react';
import { notificationService } from '../services/notificationService';
import { notificationWS } from '../services/notificationWebSocket';

export function useNotifications() {
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    loadNotifications();
    loadUnreadCount();

    // Connect to WebSocket
    const token = localStorage.getItem('access_token');
    notificationWS.connect(token);

    // Listen for new notifications
    notificationWS.onNotification((notification) => {
      setNotifications((prev) => [notification, ...prev]);
      setUnreadCount((prev) => prev + 1);
    });

    return () => {
      notificationWS.disconnect();
    };
  }, []);

  const loadNotifications = async () => {
    const data = await notificationService.getNotifications();
    setNotifications(data.notifications);
  };

  const loadUnreadCount = async () => {
    const count = await notificationService.getUnreadCount();
    setUnreadCount(count);
  };

  const markAsRead = async (id) => {
    await notificationService.markAsRead(id);
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
    );
    setUnreadCount((prev) => Math.max(0, prev - 1));
  };

  return { notifications, unreadCount, markAsRead, loadNotifications };
}
```

---

## 🎮 Live Quiz Sessions

### REST API

```javascript
// src/services/liveSessionService.js
import api from './api';

export const liveSessionService = {
  // Create session
  async createSession(sessionData) {
    const { data } = await api.post('/live-sessions', sessionData);
    return data.data;
  },

  // Get public sessions
  async getPublicSessions(status = 'upcoming') {
    const { data } = await api.get('/live-sessions/public', {
      params: { status }
    });
    return data.data;
  },

  // Get session
  async getSession(sessionId) {
    const { data } = await api.get(`/live-sessions/${sessionId}`);
    return data.data;
  },

  // Join session
  async joinSession(sessionId, accessCode = null) {
    const { data } = await api.post(`/live-sessions/${sessionId}/join`, {
      access_code: accessCode
    });
    return data.data;
  },

  // Start session (creator)
  async startSession(sessionId) {
    const { data } = await api.post(`/live-sessions/${sessionId}/start`);
    return data.data;
  },

  // Submit answer
  async submitAnswer(sessionId, questionId, answerId, responseTimeMs) {
    const { data } = await api.post(`/live-sessions/${sessionId}/answer`, {
      question_id: questionId,
      answer_id: answerId,
      response_time_ms: responseTimeMs
    });
    return data.data;
  },

  // Get leaderboard
  async getLeaderboard(sessionId) {
    const { data } = await api.get(`/live-sessions/${sessionId}/leaderboard`);
    return data.data;
  }
};
```

### WebSocket Connection

```javascript
// src/services/liveSessionWebSocket.js
class LiveSessionWebSocket {
  constructor(sessionId, token) {
    this.sessionId = sessionId;
    this.token = token;
    this.ws = null;
    this.handlers = {};
  }

  connect() {
    const wsUrl = `ws://localhost:8080/api/v1/live-sessions/${this.sessionId}/ws?token=${this.token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      const handler = this.handlers[message.type];
      if (handler) {
        handler(message.data);
      }
    };
  }

  on(eventType, handler) {
    this.handlers[eventType] = handler;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

export default LiveSessionWebSocket;
```

### React Component Example

```javascript
// src/components/LiveSession.jsx
import { useState, useEffect } from 'react';
import LiveSessionWebSocket from '../services/liveSessionWebSocket';

export function LiveSession({ sessionId }) {
  const [session, setSession] = useState(null);
  const [currentQuestion, setCurrentQuestion] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    const ws = new LiveSessionWebSocket(sessionId, token);

    ws.on('question_start', (data) => {
      setCurrentQuestion(data.question);
    });

    ws.on('question_end', () => {
      setCurrentQuestion(null);
    });

    ws.on('leaderboard_update', (data) => {
      setLeaderboard(data.leaderboard);
    });

    ws.connect();

    return () => ws.disconnect();
  }, [sessionId]);

  return (
    <div>
      {currentQuestion && (
        <div>
          <h2>{currentQuestion.question_text}</h2>
          {/* Render answers */}
        </div>
      )}
      <div>
        <h3>Leaderboard</h3>
        {leaderboard.map((entry, index) => (
          <div key={entry.user_id}>
            {index + 1}. {entry.username} - {entry.score}
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## 👨‍🏫 Educators

```javascript
// src/services/educatorService.js
import api from './api';

export const educatorService = {
  // List educators (public)
  async listEducators(params = {}) {
    const { data } = await api.get('/educators', { params });
    return data.data;
  },

  // Get educator profile
  async getEducator(educatorId) {
    const { data } = await api.get(`/educators/${educatorId}`);
    return data.data;
  },

  // Promote to educator
  async promoteToEducator(profileData) {
    const { data } = await api.post('/educator/promote', profileData);
    return data.data;
  },

  // Update educator profile
  async updateProfile(updates) {
    const { data } = await api.patch('/educator/profile', updates);
    return data.data;
  },

  // Follow educator
  async followEducator(educatorId) {
    const { data } = await api.post(`/follow/educators/${educatorId}/follow`);
    return data.data;
  },

  // Unfollow educator
  async unfollowEducator(educatorId) {
    await api.delete(`/follow/educators/${educatorId}/unfollow`);
  },

  // Get my following
  async getMyFollowing() {
    const { data } = await api.get('/follow/following');
    return data.data;
  }
};
```

---

## 📁 File Upload

```javascript
// src/services/fileService.js
import api from './api';

export const fileService = {
  // Upload file
  async uploadFile(file) {
    const formData = new FormData();
    formData.append('file', file);

    const { data } = await api.post('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    return data.data;
  },

  // Get file URL
  getFileUrl(fileId) {
    return `${api.defaults.baseURL}/files/serve/${fileId}`;
  },

  // Delete file
  async deleteFile(fileId) {
    await api.delete(`/files/${fileId}`);
  }
};
```

---

## 🎨 UI Components Examples

### Loading State

```javascript
function LoadingSpinner() {
  return <div className="spinner">Loading...</div>;
}
```

### Error Handling

```javascript
function ErrorMessage({ error }) {
  const getMessage = (error) => {
    if (error.response?.data?.message) {
      return error.response.data.message;
    }
    return 'An error occurred. Please try again.';
  };

  return (
    <div className="error">
      {getMessage(error)}
    </div>
  );
}
```

### Protected Route

```javascript
// src/components/ProtectedRoute.jsx
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();

  if (loading) {
    return <LoadingSpinner />;
  }

  if (!user) {
    return <Navigate to="/login" />;
  }

  return children;
}
```

---

## 🧪 Testing

### Example with Jest

```javascript
// src/services/__tests__/quizService.test.js
import { quizService } from '../quizService';
import api from '../api';

jest.mock('../api');

describe('quizService', () => {
  it('creates a quiz', async () => {
    const mockQuiz = { id: '123', title: 'Test Quiz' };
    api.post.mockResolvedValue({ data: { data: mockQuiz } });

    const result = await quizService.createQuiz({ title: 'Test Quiz' });

    expect(result).toEqual(mockQuiz);
    expect(api.post).toHaveBeenCalledWith('/quizzes', { title: 'Test Quiz' });
  });
});
```

---

## 📚 Additional Resources

- **Complete API Reference:** [COMPLETE_API_REFERENCE.md](COMPLETE_API_REFERENCE.md)
- **Quick Start Guide:** [QUICKSTART.md](QUICKSTART.md)
- **Backend README:** [../README.md](../README.md)

---

## 🆘 Common Issues

### CORS Errors
Make sure the backend allows your frontend origin in CORS settings.

### Token Expiration
Implement automatic token refresh using axios interceptors (shown above).

### WebSocket Connection Failed
Check that WebSocket URL is correct and token is valid.

### File Upload Fails
Ensure `Content-Type` is `multipart/form-data` and file size is within limits.

---

**Last Updated:** February 7, 2026
