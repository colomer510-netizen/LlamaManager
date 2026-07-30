using System;
using System.Diagnostics;
using System.IO;
using System.Reflection;

namespace LlamaManagerLauncher
{
    class Program
    {
        static void Main(string[] args)
        {
            string exePath = Assembly.GetExecutingAssembly().Location;
            string exeDir = Path.GetDirectoryName(exePath);
            string scriptPath = Path.Combine(exeDir, "LlamaManager.ps1");
            
            if (!File.Exists(scriptPath)) {
                Console.WriteLine("Error: No se encontro LlamaManager.ps1 en la misma carpeta.");
                Console.WriteLine("Por favor asegurate de que LlamaManager.exe y LlamaManager.ps1 esten juntos.");
                Console.WriteLine("Presiona cualquier tecla para salir...");
                Console.ReadKey();
                return;
            }

            ProcessStartInfo psi = new ProcessStartInfo();
            psi.FileName = "powershell.exe";
            psi.Arguments = string.Format("-ExecutionPolicy Bypass -NoProfile -File \"{0}\"", scriptPath);
            psi.UseShellExecute = false;
            
            Process p = Process.Start(psi);
            p.WaitForExit();
        }
    }
}
