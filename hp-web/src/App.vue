<template>
  <router-view></router-view>
</template>
<style>

.btn {
  min-width: 100px;
  height: 34px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s ease;
}

.btn.delete {
  color: #ffffff !important;
  background-color: #ef4444 !important;
  border: 1px solid #ef4444 !important;
}

.btn.delete:hover {
  color: #ffffff;
  background-color: #dc2626 !important;
  border-color: #dc2626 !important;
}
.btn.delete:active {
  color: #ffffff !important;
  background-color: #dc2626 !important;
  border-color: #dc2626 !important;
}
.btn.stop {
  color: #f97316 !important;
  background-color: #fff7ed !important;
  border: 1px solid #fdba74 !important;
}

.btn.stop:hover {
  color: #f97316 !important;
  background-color: #ffedd5 !important;
  border-color: #f97316 !important;
}
.btn.stop:active {
  color: #f97316 !important;
  background-color: #ffedd5 !important;
  border-color: #f97316 !important;
}

.btn.view {
  color: #3b82f6 !important;
  background-color: #eff6ff !important;
  border: 1px solid #bfdbfe !important;
}

.btn.view:hover {
  color: #3b82f6 !important;
  background-color: #dbeafe !important;
  border-color: #93c5fd !important;
}

.btn.view:active {
  color: #3b82f6 !important;
  background-color: #dbeafe !important;
  border-color: #93c5fd !important;
}


.btn.edit {
  color: #10b981 !important;
  background-color: #ecfdf5 !important;
  border: 1px solid #a7f3d0 !important;
}

.btn.edit:hover {
  color: #10b981 !important;
  background-color: #d1fae5 !important;
  border-color: #6ee7b7 !important;
}
.btn.edit:active {
  color: #10b981 !important;
  background-color: #d1fae5 !important;
  border-color: #6ee7b7 !important;
}

</style>
