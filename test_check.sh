if ! proot-distro list 2>/dev/null | grep -q "ubuntu.*installed"; then
  echo "would install"
fi
