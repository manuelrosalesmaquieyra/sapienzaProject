<template>
    <div class="login-container">
      <div class="login-box">
        <h1>WASAText</h1>
        <slot></slot>
      </div>
    </div>
  </template>
  
  <style scoped>
  .login-container {
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #f5f5f5;
  }
  
  .login-box {
    width: 100%;
    max-width: 400px;
    padding: 2rem;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }
  </style>