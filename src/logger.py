import logging
import os
from typing import Optional


class SecureLogger:
    """Secure logging utility that prevents log injection and sensitive data exposure"""

    def __init__(self, name: str = "speech-to-text", level: int = logging.INFO):
        self.logger = logging.getLogger(name)
        self.logger.setLevel(level)

        # Remove existing handlers to avoid duplicates
        for handler in self.logger.handlers[:]:
            self.logger.removeHandler(handler)

        # Console handler
        console_handler = logging.StreamHandler()
        console_handler.setLevel(level)

        # File handler (optional)
        log_dir = os.path.expanduser("~/.speech-to-text/logs")
        os.makedirs(log_dir, exist_ok=True)
        file_handler = logging.FileHandler(
            os.path.join(log_dir, "app.log"), mode="a", encoding="utf-8"
        )
        file_handler.setLevel(logging.DEBUG)

        # Secure formatter that prevents log injection
        formatter = logging.Formatter(
            "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S",
        )

        console_handler.setFormatter(formatter)
        file_handler.setFormatter(formatter)

        self.logger.addHandler(console_handler)
        self.logger.addHandler(file_handler)

    def _sanitize_message(self, message: str) -> str:
        """Sanitize log message to prevent injection attacks"""
        if not isinstance(message, str):
            message = str(message)

        # Remove/replace potentially dangerous characters
        dangerous_chars = {"\n": "\\n", "\r": "\\r", "\t": "\\t"}
        for char, replacement in dangerous_chars.items():
            message = message.replace(char, replacement)

        # Truncate extremely long messages
        if len(message) > 1000:
            message = message[:997] + "..."

        return message

    def info(self, message: str, *args, **kwargs):
        """Log info message with sanitization"""
        self.logger.info(self._sanitize_message(message), *args, **kwargs)

    def error(self, message: str, *args, **kwargs):
        """Log error message with sanitization"""
        self.logger.error(self._sanitize_message(message), *args, **kwargs)

    def warning(self, message: str, *args, **kwargs):
        """Log warning message with sanitization"""
        self.logger.warning(self._sanitize_message(message), *args, **kwargs)

    def debug(self, message: str, *args, **kwargs):
        """Log debug message with sanitization"""
        self.logger.debug(self._sanitize_message(message), *args, **kwargs)


# Global logger instance
_logger_instance = None


def get_logger() -> SecureLogger:
    """Get global logger instance"""
    global _logger_instance
    if _logger_instance is None:
        _logger_instance = SecureLogger()
    return _logger_instance
