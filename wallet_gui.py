#!/usr/bin/env python3
"""
Obsidian GUI Wallet

A beautiful graphical wallet for the Obsidian cryptocurrency supporting both
transparent (obs) and shielded (zobs) addresses.
"""

import tkinter as tk
from tkinter import ttk, messagebox, scrolledtext, filedialog
import threading
import time
from wallet import ObsidianWallet


class ObsidianGUIWallet:
    def __init__(self, root):
        self.root = root
        self.root.title("Obsidian Wallet")
        self.root.geometry("1000x700")
        self.root.configure(bg="#1a1a2e")
        
        # Initialize wallet
        self.wallet = None
        self.wallet_file = "gui_wallet.json"
        
        # Configure styles
        self.setup_styles()
        
        # Create main container
        self.main_container = tk.Frame(root, bg="#1a1a2e")
        self.main_container.pack(fill=tk.BOTH, expand=True, padx=20, pady=20)
        
        # Show initial screen
        self.show_welcome_screen()
        
    def setup_styles(self):
        """Setup custom styles for the GUI."""
        style = ttk.Style()
        style.theme_use('clam')
        
        # Configure colors
        bg_color = "#1a1a2e"
        fg_color = "#eee"
        accent_color = "#16213e"
        button_color = "#0f3460"
        
        style.configure("TFrame", background=bg_color)
        style.configure("TLabel", background=bg_color, foreground=fg_color, font=("Helvetica", 11))
        style.configure("Title.TLabel", font=("Helvetica", 24, "bold"), foreground="#00d4ff")
        style.configure("Header.TLabel", font=("Helvetica", 16, "bold"), foreground="#00d4ff")
        style.configure("TButton", background=button_color, foreground=fg_color, borderwidth=0, font=("Helvetica", 11))
        style.map("TButton", background=[("active", "#1a5490")])
        style.configure("TEntry", fieldbackground=accent_color, foreground=fg_color, borderwidth=2)
        
    def clear_screen(self):
        """Clear the main container."""
        for widget in self.main_container.winfo_children():
            widget.destroy()
            
    def show_welcome_screen(self):
        """Show welcome screen with options to create or load wallet."""
        self.clear_screen()
        
        # Title
        title = ttk.Label(self.main_container, text="🌑 Obsidian Wallet", style="Title.TLabel")
        title.pack(pady=50)
        
        subtitle = ttk.Label(self.main_container, text="Privacy-First Cryptocurrency Wallet", 
                            font=("Helvetica", 12))
        subtitle.pack(pady=10)
        
        # Buttons frame
        buttons_frame = ttk.Frame(self.main_container)
        buttons_frame.pack(pady=50)
        
        create_btn = tk.Button(buttons_frame, text="Create New Wallet", 
                               command=self.show_create_wallet,
                               bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 14, "bold"),
                               padx=30, pady=15, borderwidth=0, cursor="hand2")
        create_btn.pack(pady=10)
        
        load_btn = tk.Button(buttons_frame, text="Load Existing Wallet", 
                            command=self.show_load_wallet,
                            bg="#0f3460", fg="#eee", font=("Helvetica", 14, "bold"),
                            padx=30, pady=15, borderwidth=0, cursor="hand2")
        load_btn.pack(pady=10)
        
        # Connection settings
        settings_frame = ttk.Frame(self.main_container)
        settings_frame.pack(pady=30)
        
        ttk.Label(settings_frame, text="RPC Connection:").grid(row=0, column=0, sticky=tk.W, pady=5)
        self.rpc_host_var = tk.StringVar(value="http://localhost:8545")
        rpc_entry = tk.Entry(settings_frame, textvariable=self.rpc_host_var, 
                            width=30, bg="#16213e", fg="#eee", font=("Helvetica", 11))
        rpc_entry.grid(row=0, column=1, padx=10, pady=5)
        
    def show_create_wallet(self):
        """Show screen to create a new wallet."""
        self.clear_screen()
        
        ttk.Label(self.main_container, text="Create New Wallet", style="Header.TLabel").pack(pady=20)
        
        info = ttk.Label(self.main_container, 
                        text="A new wallet will be created with a 24-word recovery phrase.\n"
                             "IMPORTANT: Write down your recovery phrase and store it safely!",
                        font=("Helvetica", 10), foreground="#ff6b6b")
        info.pack(pady=20)
        
        create_btn = tk.Button(self.main_container, text="Generate Wallet", 
                              command=self.create_new_wallet,
                              bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 12, "bold"),
                              padx=20, pady=10, cursor="hand2")
        create_btn.pack(pady=20)
        
        back_btn = tk.Button(self.main_container, text="← Back", 
                            command=self.show_welcome_screen,
                            bg="#16213e", fg="#eee", font=("Helvetica", 10),
                            padx=15, pady=5, cursor="hand2")
        back_btn.pack(pady=10)
        
    def create_new_wallet(self):
        """Create a new wallet and show the mnemonic."""
        try:
            self.wallet = ObsidianWallet(
                rpc_host=self.rpc_host_var.get(),
                wallet_file=self.wallet_file,
                test_mode=False
            )
            
            mnemonic = self.wallet.generate_mnemonic()
            
            # Generate first addresses
            t_addr = self.wallet.generate_transparent_address()
            z_addr = self.wallet.generate_shielded_address()
            
            # Save wallet
            self.wallet.save_wallet(self.wallet_file)
            
            # Show mnemonic
            self.show_mnemonic(mnemonic)
            
        except Exception as e:
            messagebox.showerror("Error", f"Failed to create wallet: {str(e)}")
            
    def show_mnemonic(self, mnemonic):
        """Display the mnemonic phrase."""
        self.clear_screen()
        
        ttk.Label(self.main_container, text="⚠️ Recovery Phrase", style="Header.TLabel").pack(pady=20)
        
        warning = ttk.Label(self.main_container, 
                           text="Write down these 24 words in order and keep them safe.\n"
                                "Anyone with this phrase can access your funds!",
                           font=("Helvetica", 10), foreground="#ff6b6b")
        warning.pack(pady=10)
        
        # Mnemonic display
        mnemonic_frame = tk.Frame(self.main_container, bg="#16213e", padx=20, pady=20)
        mnemonic_frame.pack(pady=20, padx=40, fill=tk.X)
        
        words = mnemonic.split()
        for i, word in enumerate(words, 1):
            row = (i - 1) // 4
            col = (i - 1) % 4
            word_label = tk.Label(mnemonic_frame, text=f"{i}. {word}", 
                                 bg="#16213e", fg="#00d4ff", 
                                 font=("Courier", 11, "bold"))
            word_label.grid(row=row, column=col, padx=15, pady=5, sticky=tk.W)
        
        # Copy button
        def copy_mnemonic():
            self.root.clipboard_clear()
            self.root.clipboard_append(mnemonic)
            messagebox.showinfo("Copied", "Recovery phrase copied to clipboard!")
            
        copy_btn = tk.Button(self.main_container, text="📋 Copy to Clipboard", 
                            command=copy_mnemonic,
                            bg="#0f3460", fg="#eee", font=("Helvetica", 10),
                            padx=15, pady=5, cursor="hand2")
        copy_btn.pack(pady=10)
        
        # Confirm button
        confirm_btn = tk.Button(self.main_container, text="I've Saved My Recovery Phrase →", 
                               command=self.show_main_wallet,
                               bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 12, "bold"),
                               padx=20, pady=10, cursor="hand2")
        confirm_btn.pack(pady=20)
        
    def show_load_wallet(self):
        """Show screen to load existing wallet."""
        self.clear_screen()
        
        ttk.Label(self.main_container, text="Load Existing Wallet", style="Header.TLabel").pack(pady=20)
        
        # Option 1: Load from file
        file_frame = tk.LabelFrame(self.main_container, text="Load from File", 
                                   bg="#16213e", fg="#eee", font=("Helvetica", 11, "bold"),
                                   padx=20, pady=20)
        file_frame.pack(pady=20, padx=40, fill=tk.X)
        
        def load_from_file():
            filename = filedialog.askopenfilename(
                title="Select Wallet File",
                filetypes=[("JSON files", "*.json"), ("All files", "*.*")]
            )
            if filename:
                try:
                    self.wallet = ObsidianWallet(
                        rpc_host=self.rpc_host_var.get(),
                        wallet_file=filename,
                        test_mode=False
                    )
                    self.wallet_file = filename
                    self.show_main_wallet()
                except Exception as e:
                    messagebox.showerror("Error", f"Failed to load wallet: {str(e)}")
        
        file_btn = tk.Button(file_frame, text="📁 Browse for Wallet File", 
                            command=load_from_file,
                            bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 11),
                            padx=15, pady=8, cursor="hand2")
        file_btn.pack()
        
        # Option 2: Restore from mnemonic
        mnemonic_frame = tk.LabelFrame(self.main_container, text="Restore from Recovery Phrase", 
                                       bg="#16213e", fg="#eee", font=("Helvetica", 11, "bold"),
                                       padx=20, pady=20)
        mnemonic_frame.pack(pady=20, padx=40, fill=tk.BOTH, expand=True)
        
        ttk.Label(mnemonic_frame, text="Enter your 24-word recovery phrase:").pack(pady=10)
        
        mnemonic_text = scrolledtext.ScrolledText(mnemonic_frame, height=4, width=60,
                                                  bg="#0f3460", fg="#eee", 
                                                  font=("Courier", 10))
        mnemonic_text.pack(pady=10)
        
        def restore_from_mnemonic():
            mnemonic = mnemonic_text.get("1.0", tk.END).strip()
            if not mnemonic:
                messagebox.showwarning("Warning", "Please enter your recovery phrase")
                return
                
            try:
                self.wallet = ObsidianWallet(
                    rpc_host=self.rpc_host_var.get(),
                    wallet_file=self.wallet_file,
                    test_mode=False
                )
                
                if self.wallet.load_mnemonic(mnemonic):
                    # Regenerate first addresses
                    self.wallet.generate_transparent_address()
                    self.wallet.generate_shielded_address()
                    self.wallet.save_wallet(self.wallet_file)
                    self.show_main_wallet()
                else:
                    messagebox.showerror("Error", "Invalid recovery phrase")
            except Exception as e:
                messagebox.showerror("Error", f"Failed to restore wallet: {str(e)}")
        
        restore_btn = tk.Button(mnemonic_frame, text="Restore Wallet", 
                               command=restore_from_mnemonic,
                               bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 11),
                               padx=15, pady=8, cursor="hand2")
        restore_btn.pack(pady=10)
        
        # Back button
        back_btn = tk.Button(self.main_container, text="← Back", 
                            command=self.show_welcome_screen,
                            bg="#16213e", fg="#eee", font=("Helvetica", 10),
                            padx=15, pady=5, cursor="hand2")
        back_btn.pack(pady=10)
        
    def show_main_wallet(self):
        """Show main wallet interface."""
        self.clear_screen()
        
        # Create notebook for tabs
        notebook = ttk.Notebook(self.main_container)
        notebook.pack(fill=tk.BOTH, expand=True)
        
        # Overview Tab
        overview_frame = tk.Frame(notebook, bg="#1a1a2e")
        notebook.add(overview_frame, text="Overview")
        self.create_overview_tab(overview_frame)
        
        # Send Tab
        send_frame = tk.Frame(notebook, bg="#1a1a2e")
        notebook.add(send_frame, text="Send")
        self.create_send_tab(send_frame)
        
        # Receive Tab
        receive_frame = tk.Frame(notebook, bg="#1a1a2e")
        notebook.add(receive_frame, text="Receive")
        self.create_receive_tab(receive_frame)
        
        # History Tab
        history_frame = tk.Frame(notebook, bg="#1a1a2e")
        notebook.add(history_frame, text="History")
        self.create_history_tab(history_frame)
        
        # Settings Tab
        settings_frame = tk.Frame(notebook, bg="#1a1a2e")
        notebook.add(settings_frame, text="Settings")
        self.create_settings_tab(settings_frame)
        
    def create_overview_tab(self, parent):
        """Create the overview tab."""
        # Balance section
        balance_frame = tk.LabelFrame(parent, text="💰 Total Balance", 
                                      bg="#16213e", fg="#eee", font=("Helvetica", 12, "bold"),
                                      padx=20, pady=20)
        balance_frame.pack(pady=20, padx=20, fill=tk.X)
        
        self.balance_label = tk.Label(balance_frame, text="Loading...", 
                                      bg="#16213e", fg="#00d4ff", 
                                      font=("Helvetica", 32, "bold"))
        self.balance_label.pack()
        
        tk.Label(balance_frame, text="OBS", bg="#16213e", fg="#999", 
                font=("Helvetica", 14)).pack()
        
        # Addresses section
        addresses_frame = tk.LabelFrame(parent, text="📍 Your Addresses", 
                                        bg="#16213e", fg="#eee", font=("Helvetica", 12, "bold"),
                                        padx=20, pady=20)
        addresses_frame.pack(pady=20, padx=20, fill=tk.BOTH, expand=True)
        
        # Create treeview for addresses
        columns = ("Type", "Address", "Balance")
        self.addresses_tree = ttk.Treeview(addresses_frame, columns=columns, show="headings", height=8)
        
        for col in columns:
            self.addresses_tree.heading(col, text=col)
            
        self.addresses_tree.column("Type", width=100)
        self.addresses_tree.column("Address", width=400)
        self.addresses_tree.column("Balance", width=150)
        
        scrollbar = ttk.Scrollbar(addresses_frame, orient=tk.VERTICAL, command=self.addresses_tree.yview)
        self.addresses_tree.configure(yscroll=scrollbar.set)
        
        self.addresses_tree.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        
        # Refresh button
        refresh_btn = tk.Button(parent, text="🔄 Refresh Balances", 
                               command=self.refresh_balances,
                               bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 11),
                               padx=15, pady=8, cursor="hand2")
        refresh_btn.pack(pady=10)
        
        # Initial load
        self.refresh_balances()
        
    def refresh_balances(self):
        """Refresh all balances."""
        def update():
            try:
                # Clear existing items
                for item in self.addresses_tree.get_children():
                    self.addresses_tree.delete(item)
                
                total_balance = 0.0
                
                # Get transparent addresses
                for priv, pub, addr in self.wallet.transparent_keys:
                    try:
                        balance_info = self.wallet.get_balance(addr)
                        balance = balance_info if isinstance(balance_info, float) else 0.0
                        self.addresses_tree.insert("", tk.END, values=("Transparent", addr, f"{balance:.8f} OBS"))
                        total_balance += balance
                    except:
                        self.addresses_tree.insert("", tk.END, values=("Transparent", addr, "Error"))
                
                # Get shielded addresses
                for pub, view, addr in self.wallet.shielded_keys:
                    try:
                        balance_info = self.wallet.get_balance(addr)
                        balance = balance_info if isinstance(balance_info, float) else 0.0
                        self.addresses_tree.insert("", tk.END, values=("Shielded", addr, f"{balance:.8f} OBS"))
                        total_balance += balance
                    except:
                        self.addresses_tree.insert("", tk.END, values=("Shielded", addr, "Error"))
                
                # Update total balance
                self.balance_label.config(text=f"{total_balance:.8f}")
                
            except Exception as e:
                messagebox.showerror("Error", f"Failed to refresh balances: {str(e)}")
        
        # Run in background thread
        threading.Thread(target=update, daemon=True).start()
        
    def create_send_tab(self, parent):
        """Create the send tab."""
        send_form = tk.Frame(parent, bg="#1a1a2e")
        send_form.pack(pady=40, padx=40, fill=tk.BOTH, expand=True)
        
        # From address
        tk.Label(send_form, text="From Address:", bg="#1a1a2e", fg="#eee", 
                font=("Helvetica", 11)).grid(row=0, column=0, sticky=tk.W, pady=10)
        
        self.send_from_var = tk.StringVar()
        from_combo = ttk.Combobox(send_form, textvariable=self.send_from_var, width=50)
        from_combo['values'] = self.get_all_addresses()
        if from_combo['values']:
            from_combo.current(0)
        from_combo.grid(row=0, column=1, pady=10, padx=10)
        
        # To address
        tk.Label(send_form, text="To Address:", bg="#1a1a2e", fg="#eee", 
                font=("Helvetica", 11)).grid(row=1, column=0, sticky=tk.W, pady=10)
        
        self.send_to_var = tk.StringVar()
        to_entry = tk.Entry(send_form, textvariable=self.send_to_var, width=53,
                           bg="#16213e", fg="#eee", font=("Helvetica", 11))
        to_entry.grid(row=1, column=1, pady=10, padx=10)
        
        # Amount
        tk.Label(send_form, text="Amount (OBS):", bg="#1a1a2e", fg="#eee", 
                font=("Helvetica", 11)).grid(row=2, column=0, sticky=tk.W, pady=10)
        
        self.send_amount_var = tk.StringVar()
        amount_entry = tk.Entry(send_form, textvariable=self.send_amount_var, width=53,
                               bg="#16213e", fg="#eee", font=("Helvetica", 11))
        amount_entry.grid(row=2, column=1, pady=10, padx=10)
        
        # Memo (optional for shielded)
        tk.Label(send_form, text="Memo (optional):", bg="#1a1a2e", fg="#eee", 
                font=("Helvetica", 11)).grid(row=3, column=0, sticky=tk.W, pady=10)
        
        self.send_memo_var = tk.StringVar()
        memo_entry = tk.Entry(send_form, textvariable=self.send_memo_var, width=53,
                             bg="#16213e", fg="#eee", font=("Helvetica", 11))
        memo_entry.grid(row=3, column=1, pady=10, padx=10)
        
        # Send button
        def send_transaction():
            from_addr = self.send_from_var.get()
            to_addr = self.send_to_var.get()
            amount_str = self.send_amount_var.get()
            memo = self.send_memo_var.get()
            
            if not from_addr or not to_addr or not amount_str:
                messagebox.showwarning("Warning", "Please fill in all required fields")
                return
            
            try:
                amount = float(amount_str)
                if amount <= 0:
                    messagebox.showwarning("Warning", "Amount must be positive")
                    return
                
                # Confirm transaction
                if not messagebox.askyesno("Confirm", 
                                          f"Send {amount} OBS from\n{from_addr}\nto\n{to_addr}?"):
                    return
                
                # Send transaction
                txid = self.wallet.send_transaction(from_addr, to_addr, amount, memo)
                messagebox.showinfo("Success", f"Transaction sent!\nTXID: {txid[:16]}...")
                
                # Clear form
                self.send_to_var.set("")
                self.send_amount_var.set("")
                self.send_memo_var.set("")
                
                # Refresh balances
                self.refresh_balances()
                
            except Exception as e:
                messagebox.showerror("Error", f"Failed to send transaction: {str(e)}")
        
        send_btn = tk.Button(send_form, text="📤 Send Transaction", 
                            command=send_transaction,
                            bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 12, "bold"),
                            padx=30, pady=12, cursor="hand2")
        send_btn.grid(row=4, column=0, columnspan=2, pady=30)
        
    def create_receive_tab(self, parent):
        """Create the receive tab."""
        receive_frame = tk.Frame(parent, bg="#1a1a2e")
        receive_frame.pack(pady=40, padx=40)
        
        ttk.Label(receive_frame, text="Receive OBS", style="Header.TLabel").pack(pady=20)
        
        # Address type selection
        type_frame = tk.Frame(receive_frame, bg="#1a1a2e")
        type_frame.pack(pady=20)
        
        self.receive_type_var = tk.StringVar(value="transparent")
        
        tk.Radiobutton(type_frame, text="Transparent Address (obs...)", 
                      variable=self.receive_type_var, value="transparent",
                      bg="#1a1a2e", fg="#eee", selectcolor="#16213e",
                      font=("Helvetica", 11)).pack(anchor=tk.W, pady=5)
        
        tk.Radiobutton(type_frame, text="Shielded Address (zobs...) - Private", 
                      variable=self.receive_type_var, value="shielded",
                      bg="#1a1a2e", fg="#eee", selectcolor="#16213e",
                      font=("Helvetica", 11)).pack(anchor=tk.W, pady=5)
        
        # Generate new address button
        def generate_address():
            addr_type = self.receive_type_var.get()
            try:
                if addr_type == "transparent":
                    addr = self.wallet.generate_transparent_address()
                else:
                    addr = self.wallet.generate_shielded_address()
                
                self.wallet.save_wallet(self.wallet_file)
                
                # Display address
                self.receive_addr_var.set(addr)
                messagebox.showinfo("Success", f"New {addr_type} address generated!")
                
                # Refresh addresses list
                self.refresh_balances()
                
            except Exception as e:
                messagebox.showerror("Error", f"Failed to generate address: {str(e)}")
        
        gen_btn = tk.Button(receive_frame, text="➕ Generate New Address", 
                           command=generate_address,
                           bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 11),
                           padx=20, pady=10, cursor="hand2")
        gen_btn.pack(pady=20)
        
        # Display address
        ttk.Label(receive_frame, text="Your Address:").pack(pady=10)
        
        self.receive_addr_var = tk.StringVar()
        addr_entry = tk.Entry(receive_frame, textvariable=self.receive_addr_var, width=60,
                             bg="#16213e", fg="#00d4ff", font=("Courier", 12, "bold"),
                             state="readonly", justify=tk.CENTER)
        addr_entry.pack(pady=10)
        
        # Copy button
        def copy_address():
            addr = self.receive_addr_var.get()
            if addr:
                self.root.clipboard_clear()
                self.root.clipboard_append(addr)
                messagebox.showinfo("Copied", "Address copied to clipboard!")
        
        copy_btn = tk.Button(receive_frame, text="📋 Copy Address", 
                            command=copy_address,
                            bg="#0f3460", fg="#eee", font=("Helvetica", 11),
                            padx=20, pady=8, cursor="hand2")
        copy_btn.pack(pady=10)
        
    def create_history_tab(self, parent):
        """Create the transaction history tab."""
        history_frame = tk.Frame(parent, bg="#1a1a2e")
        history_frame.pack(pady=20, padx=20, fill=tk.BOTH, expand=True)
        
        ttk.Label(history_frame, text="Transaction History", style="Header.TLabel").pack(pady=10)
        
        # Create treeview for history
        columns = ("Time", "Type", "From", "To", "Amount", "TXID")
        self.history_tree = ttk.Treeview(history_frame, columns=columns, show="headings", height=15)
        
        for col in columns:
            self.history_tree.heading(col, text=col)
        
        self.history_tree.column("Time", width=150)
        self.history_tree.column("Type", width=100)
        self.history_tree.column("From", width=180)
        self.history_tree.column("To", width=180)
        self.history_tree.column("Amount", width=120)
        self.history_tree.column("TXID", width=200)
        
        scrollbar = ttk.Scrollbar(history_frame, orient=tk.VERTICAL, command=self.history_tree.yview)
        self.history_tree.configure(yscroll=scrollbar.set)
        
        self.history_tree.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        
        # Load history
        self.refresh_history()
        
        # Refresh button
        refresh_btn = tk.Button(parent, text="🔄 Refresh History", 
                               command=self.refresh_history,
                               bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 11),
                               padx=15, pady=8, cursor="hand2")
        refresh_btn.pack(pady=10)
        
    def refresh_history(self):
        """Refresh transaction history."""
        # Clear existing items
        for item in self.history_tree.get_children():
            self.history_tree.delete(item)
        
        # Load history from wallet
        for tx in self.wallet.transaction_history:
            timestamp = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(tx['timestamp']))
            tx_type = tx.get('type', 'Unknown')
            from_addr = tx['from'][:20] + "..." if len(tx['from']) > 20 else tx['from']
            to_addr = tx['to'][:20] + "..." if len(tx['to']) > 20 else tx['to']
            amount = f"{tx['amount']:.8f} OBS"
            txid = tx['txid'][:16] + "..." if len(tx['txid']) > 16 else tx['txid']
            
            self.history_tree.insert("", tk.END, values=(timestamp, tx_type, from_addr, to_addr, amount, txid))
        
    def create_settings_tab(self, parent):
        """Create the settings tab."""
        settings_frame = tk.Frame(parent, bg="#1a1a2e")
        settings_frame.pack(pady=40, padx=40, fill=tk.BOTH, expand=True)
        
        ttk.Label(settings_frame, text="Settings", style="Header.TLabel").pack(pady=20)
        
        # RPC settings
        rpc_frame = tk.LabelFrame(settings_frame, text="RPC Connection", 
                                  bg="#16213e", fg="#eee", font=("Helvetica", 11, "bold"),
                                  padx=20, pady=20)
        rpc_frame.pack(pady=20, fill=tk.X)
        
        tk.Label(rpc_frame, text="RPC Host:", bg="#16213e", fg="#eee", 
                font=("Helvetica", 11)).grid(row=0, column=0, sticky=tk.W, pady=10)
        
        rpc_entry = tk.Entry(rpc_frame, textvariable=self.rpc_host_var, width=40,
                            bg="#0f3460", fg="#eee", font=("Helvetica", 11))
        rpc_entry.grid(row=0, column=1, pady=10, padx=10)
        
        # Wallet info
        info_frame = tk.LabelFrame(settings_frame, text="Wallet Information", 
                                   bg="#16213e", fg="#eee", font=("Helvetica", 11, "bold"),
                                   padx=20, pady=20)
        info_frame.pack(pady=20, fill=tk.X)
        
        tk.Label(info_frame, text=f"Wallet File: {self.wallet_file}", 
                bg="#16213e", fg="#eee", font=("Helvetica", 10)).pack(anchor=tk.W, pady=5)
        
        tk.Label(info_frame, text=f"Transparent Addresses: {len(self.wallet.transparent_keys)}", 
                bg="#16213e", fg="#eee", font=("Helvetica", 10)).pack(anchor=tk.W, pady=5)
        
        tk.Label(info_frame, text=f"Shielded Addresses: {len(self.wallet.shielded_keys)}", 
                bg="#16213e", fg="#eee", font=("Helvetica", 10)).pack(anchor=tk.W, pady=5)
        
        # Backup wallet
        def backup_wallet():
            filename = filedialog.asksaveasfilename(
                defaultextension=".json",
                filetypes=[("JSON files", "*.json"), ("All files", "*.*")],
                initialfile="obsidian_wallet_backup.json"
            )
            if filename:
                try:
                    self.wallet.save_wallet(filename)
                    messagebox.showinfo("Success", f"Wallet backed up to {filename}")
                except Exception as e:
                    messagebox.showerror("Error", f"Failed to backup wallet: {str(e)}")
        
        backup_btn = tk.Button(settings_frame, text="💾 Backup Wallet", 
                              command=backup_wallet,
                              bg="#00d4ff", fg="#1a1a2e", font=("Helvetica", 11),
                              padx=20, pady=10, cursor="hand2")
        backup_btn.pack(pady=20)
        
        # Show recovery phrase
        def show_recovery():
            if messagebox.askyesno("Warning", "This will display your recovery phrase. Make sure no one is watching!"):
                messagebox.showinfo("Recovery Phrase", self.wallet.mnemonic)
        
        recovery_btn = tk.Button(settings_frame, text="🔑 Show Recovery Phrase", 
                                command=show_recovery,
                                bg="#ff6b6b", fg="#fff", font=("Helvetica", 11),
                                padx=20, pady=10, cursor="hand2")
        recovery_btn.pack(pady=10)
        
    def get_all_addresses(self):
        """Get all wallet addresses."""
        addresses = []
        for _, _, addr in self.wallet.transparent_keys:
            addresses.append(addr)
        for _, _, addr in self.wallet.shielded_keys:
            addresses.append(addr)
        return addresses


def main():
    root = tk.Tk()
    app = ObsidianGUIWallet(root)
    root.mainloop()


if __name__ == "__main__":
    main()
